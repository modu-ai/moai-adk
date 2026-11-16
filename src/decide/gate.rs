//! The gate of `moai decide` (card C-37, design v2 slice V1): which choices a model may make
//! at all.
//!
//! SKILL.md sorts an open choice: one that cannot be undone, that touches one of the five HOLD
//! categories or that changes a `guarded` setting goes to the user and waits. Only what is
//! left, a reversible choice of category `none`, is the leader's, and only that may be put to
//! a model in the leader's place. So the gate lets through exactly: `reversible` is `true`,
//! `category` is `none`, the default is one of the options, and no text of the request holds
//! a shape of the deny list. Eligibility rests on those honest structured fields (with `undo`
//! and `tier`, which the parser requires), never on a model and never on a word list.
//!
//! The deny list ([`DENY`]) is for known threats only, not a risk judge. It names a few exact
//! command and action shapes that must never be handed to a model whatever the fields say,
//! and nothing else: a force push, `reset --hard`, a forced recursive delete, dropping or
//! truncating a database table, making a repository public, `chmod 777`, and a currency
//! amount. It reads `what`, every option value and description, `undo` and every evidence
//! value, in any letter case, one line at a time. A text without one of these shapes is not
//! shown to be harmless; that judgement stays with the person who filled in `reversible`,
//! `category`, `undo` and `tier`. Each shape refuses with its own code, before any provider
//! code runs.
//!
//! The first design's keyword screen (a word list over the same texts, with a
//! `class_mismatch` refusal) is gone: it let consequential sentences through and refused
//! ordinary ones such as "Remove the unused import?" (design v2, two audits of the first
//! design).

use super::request::Request;
use super::Refusal;
use crate::decision_file::CATEGORIES;

/// A request the gate let through. Made only by [`check`], and the only thing
/// [`super::outbound::build`] accepts.
#[derive(Debug)]
pub struct Cleared<'a> {
    request: &'a Request,
}

impl<'a> Cleared<'a> {
    pub fn request(&self) -> &'a Request {
        self.request
    }

    /// For the tests of the steps after the gate, which must not depend on its rules.
    #[cfg(test)]
    pub(crate) fn unchecked(request: &'a Request) -> Self {
        Cleared { request }
    }
}

/// The deny list: each shape's reason code and what it names. Known threats only, not a risk
/// judge (see the module text); [`deny`] says how each one is matched.
pub const DENY: [(&str, &str); 7] = [
    (
        "deny_force_push",
        "a force push (`push` with `--force` or `-f` after it, or `--force-with-lease`)",
    ),
    ("deny_reset_hard", "`reset` with `--hard` after it"),
    (
        "deny_rm_rf",
        "a forced recursive delete (`rm` with both the `r` and the `f` flag after it)",
    ),
    (
        "deny_drop_table",
        "`drop table`, `drop database` or `truncate table`",
    ),
    (
        "deny_visibility_public",
        "making a repository public (`--visibility public`, `visibility=public`)",
    ),
    ("deny_chmod_777", "`chmod` with `777` or `0777` after it"),
    (
        "deny_currency",
        "a currency amount (`$`, `€`, `£`, `₩` or `¥` right before a digit, or a number before \
         USD, EUR, KRW or 원)",
    ),
];

/// The code of the first deny-list shape `text` holds, or `None`.
///
/// The command shapes are read one line at a time, as words split at white space, in lower
/// case, with quotes, brackets and sentence punctuation taken off each word's ends:
/// - `deny_force_push`: the word `push` with the word `--force` or `-f` anywhere after it on
///   the line (`git push origin main --force` too), or the word `--force-with-lease` (also
///   `--force-with-lease=…`) anywhere. `--force` without `push` before it is not this shape.
/// - `deny_reset_hard`: the word `reset` with `--hard` after it on the line.
/// - `deny_rm_rf`: the word `rm` with both the `r` and the `f` flag after it on the line, in
///   one word (`-rf`, `-fr`, `-Rf`, `-rfv`), in two (`-r -f`) or long (`--recursive
///   --force`). `rm -r` alone passes: a recursive delete without force is judged by the
///   honest `reversible` and `undo` fields, not by this list.
/// - `deny_drop_table`: the two words `drop table`, `drop database` or `truncate table`.
/// - `deny_visibility_public`: `visibility=public` anywhere, or the words `--visibility
///   public`.
/// - `deny_chmod_777`: the word `chmod` with the word `777` or `0777` after it on the line.
///
/// `deny_currency` reads the text as it is: `$`, `€`, `£`, `₩` or `¥` right before an ASCII
/// digit (`$5`, `₩5000`; `$HOME` and `$ 5` are not), or a number (digits, with `,` or `.`
/// inside) right before `USD`, `EUR` or `KRW` as a whole word, in any case and with at most
/// one space between, or right before `원`. Known false hits, each costing one fallback to
/// the leader: a shell argument such as `$1`, and a Korean word that starts with `원` after
/// a number (`제1원칙`).
pub fn deny(text: &str) -> Option<&'static str> {
    for line in text.lines() {
        let words = words(line);
        let after = |first: &str, found: &dyn Fn(&str) -> bool| {
            words
                .iter()
                .position(|w| w == first)
                .is_some_and(|at| words[at + 1..].iter().any(|w| found(w)))
        };
        let pair = |a: &str, b: &str| words.windows(2).any(|w| w[0] == a && w[1] == b);
        if after("push", &|w| w == "--force" || w == "-f")
            || words
                .iter()
                .any(|w| w == "--force-with-lease" || w.starts_with("--force-with-lease="))
        {
            return Some("deny_force_push");
        }
        if after("reset", &|w| w == "--hard") {
            return Some("deny_reset_hard");
        }
        if rm_forced_recursive(&words) {
            return Some("deny_rm_rf");
        }
        if pair("drop", "table") || pair("drop", "database") || pair("truncate", "table") {
            return Some("deny_drop_table");
        }
        if line.to_lowercase().contains("visibility=public") || pair("--visibility", "public") {
            return Some("deny_visibility_public");
        }
        if after("chmod", &|w| w == "777" || w == "0777") {
            return Some("deny_chmod_777");
        }
    }
    holds_currency_amount(text).then_some("deny_currency")
}

/// The words of one line: split at white space, lower case, with quotes, brackets and
/// sentence punctuation taken off both ends.
fn words(line: &str) -> Vec<String> {
    line.split_whitespace()
        .map(|w| {
            w.trim_matches(|c: char| "`'\"()[]{}<>,;:.!?".contains(c))
                .to_lowercase()
        })
        .filter(|w| !w.is_empty())
        .collect()
}

/// True when the word `rm` is followed on its line by both the recursive and the force flag.
fn rm_forced_recursive(words: &[String]) -> bool {
    let Some(at) = words.iter().position(|w| w == "rm") else {
        return false;
    };
    let (mut recursive, mut force) = (false, false);
    for word in &words[at + 1..] {
        match word.as_str() {
            "--recursive" => recursive = true,
            "--force" => force = true,
            short if short.starts_with('-') && !short.starts_with("--") => {
                recursive |= short.contains('r');
                force |= short.contains('f');
            }
            _ => {}
        }
    }
    recursive && force
}

/// True when `text` holds a currency amount (see [`deny`]).
fn holds_currency_amount(text: &str) -> bool {
    let chars: Vec<char> = text.chars().collect();
    let symbol_before_digit = chars
        .windows(2)
        .any(|w| "$€£₩¥".contains(w[0]) && w[1].is_ascii_digit());
    if symbol_before_digit {
        return true;
    }
    let mut i = 0;
    while i < chars.len() {
        if !chars[i].is_ascii_digit() {
            i += 1;
            continue;
        }
        // The number: digits, with `,` or `.` between digits.
        while i < chars.len()
            && (chars[i].is_ascii_digit()
                || (matches!(chars[i], ',' | '.')
                    && chars.get(i + 1).is_some_and(char::is_ascii_digit)))
        {
            i += 1;
        }
        let mut at = i;
        if chars.get(at) == Some(&' ') {
            at += 1;
        }
        if chars.get(at) == Some(&'원') {
            return true;
        }
        let unit: String = chars[at..]
            .iter()
            .take(3)
            .collect::<String>()
            .to_lowercase();
        let ends = !chars.get(at + 3).is_some_and(|c| c.is_ascii_alphanumeric());
        if ["usd", "eur", "krw"].contains(&unit.as_str()) && ends {
            return true;
        }
    }
    false
}

/// The gate: `request` as a [`Cleared`] choice, or the first rule that refuses it. The order
/// is fixed: reversible, category, default, then the deny list.
pub fn check(request: &Request) -> Result<Cleared<'_>, Refusal> {
    if request.reversible != Some(true) {
        return Err(Refusal::new(
            "irreversible",
            "`reversible` is not `true`: a choice that cannot be undone waits for the user",
        ));
    }
    let Some(declared) = request.category.as_deref() else {
        return Err(Refusal::new(
            "category_missing",
            "`category` is missing: a choice without a category is not the model's to make",
        ));
    };
    // The shared name, so a refusal prints a fixed word and never the request's own text.
    let Some(category) = CATEGORIES.iter().copied().find(|c| *c == declared) else {
        return Err(Refusal::new(
            "category_unknown",
            format!("`category` is not one of {}", CATEGORIES.join(", ")),
        ));
    };
    match category {
        "none" => {}
        "guarded" => {
            return Err(Refusal::new(
                "category_guarded",
                "category `guarded`: a guarded setting waits for the user",
            ))
        }
        hold => {
            return Err(Refusal::new(
                "category_hold",
                format!("category `{hold}` is a HOLD category: it waits for the user"),
            ))
        }
    }
    if !request.options.iter().any(|o| o.value == request.default) {
        return Err(Refusal::new(
            "default_not_an_option",
            "`default` is not the value of one of the options",
        ));
    }
    // Each text with the fixed name of where it stands, so the reason never quotes the request
    // (an evidence field's name is the request's own text too).
    let texts = [("`what`", request.what.as_str())]
        .into_iter()
        .chain(request.options.iter().flat_map(|o| {
            [
                ("an option value", o.value.as_str()),
                ("an option description", o.description.as_str()),
            ]
        }))
        .chain(request.undo.as_deref().map(|undo| ("`undo`", undo)))
        .chain(
            request
                .evidence
                .values()
                .map(|value| ("an evidence value", value.as_str())),
        );
    for (place, text) in texts {
        if let Some(code) = deny(text) {
            let names = DENY
                .iter()
                .find(|(c, _)| *c == code)
                .map(|(_, names)| *names)
                .unwrap_or_default();
            return Err(Refusal::new(
                code,
                format!("{place} holds {names}: a known threat shape, the leader decides"),
            ));
        }
    }
    Ok(Cleared { request })
}

#[cfg(test)]
mod tests {
    use super::super::request::parse;
    use super::super::testing::{evidence, request_json};
    use super::*;
    use serde_json::{json, Value};

    fn code(edit: impl FnOnce(&mut Value)) -> &'static str {
        let request = parse(&request_json(edit)).unwrap();
        match check(&request) {
            Ok(_) => "cleared",
            Err(r) => r.code,
        }
    }

    #[test]
    fn only_a_reversible_none_with_its_default_among_the_options_is_cleared() {
        assert_eq!(code(|_| {}), "cleared");
        assert_eq!(code(|v| v["default"] = json!("cargo-test")), "cleared");
        assert_eq!(code(|v| v["reversible"] = json!(false)), "irreversible");
        assert_eq!(
            code(|v| v["default"] = json!("Nextest")),
            "default_not_an_option"
        );
    }

    /// Every shared category, literally, and what the gate answers; the list here and the
    /// shared one are the same seven.
    #[test]
    fn each_category_has_its_answer() {
        let table = [
            ("none", "cleared"),
            ("guarded", "category_guarded"),
            ("money", "category_hold"),
            ("legal", "category_hold"),
            ("external_contract", "category_hold"),
            ("secret", "category_hold"),
            ("real_data", "category_hold"),
        ];
        assert_eq!(table.map(|(c, _)| c), CATEGORIES);
        for (category, want) in table {
            assert_eq!(
                code(|v| v["category"] = json!(category)),
                want,
                "{category}"
            );
        }
        assert_eq!(code(|v| v["category"] = json!("hold")), "category_unknown");
        assert_eq!(
            code(|v| {
                v.as_object_mut().unwrap().remove("category");
            }),
            "category_missing"
        );
    }

    /// The first failing rule answers: a one-way choice is `irreversible` whatever else is
    /// wrong with it, a category is judged before the default, and the default before the
    /// deny list.
    #[test]
    fn the_rules_answer_in_a_fixed_order() {
        assert_eq!(
            code(|v| {
                v["reversible"] = json!(false);
                v["category"] = json!("money");
                v["default"] = json!("x");
            }),
            "irreversible"
        );
        assert_eq!(
            code(|v| {
                v["category"] = json!("secret");
                v["default"] = json!("x");
            }),
            "category_hold"
        );
        assert_eq!(
            code(|v| {
                v["default"] = json!("x");
                v["what"] = json!("Run git push --force now?");
            }),
            "default_not_an_option"
        );
    }

    /// The ordinary engineering choices the first design's keyword screen refused: with
    /// honest fields, each one is cleared.
    #[test]
    fn ordinary_choices_pass_with_honest_fields() {
        for (what, options, undo) in [
            (
                "Remove the unused import?",
                [
                    ("remove", "delete the unused use line"),
                    ("keep", "leave it"),
                ],
                "add the use line back",
            ),
            (
                "Drop the trailing comma in the fixture?",
                [("drop", "drop the comma"), ("keep", "keep the comma")],
                "put the comma back",
            ),
            (
                "Delete the dead code path?",
                [
                    ("delete", "delete the unreachable branch"),
                    ("keep", "keep it"),
                ],
                "git revert the commit that deletes it",
            ),
            (
                "Rotate the log files daily or weekly?",
                [
                    ("daily", "rotate every day"),
                    ("weekly", "rotate every week"),
                ],
                "change the rotation setting back",
            ),
        ] {
            let c = code(|v| {
                v["what"] = json!(what);
                v["options"] = json!([
                    {"value": options[0].0, "description": options[0].1},
                    {"value": options[1].0, "description": options[1].1},
                ]);
                v["default"] = json!(options[1].0);
                v["undo"] = json!(undo);
                v["evidence"] = evidence("the file has one such line");
            });
            assert_eq!(c, "cleared", "{what}");
        }
    }

    /// Each deny shape, refused with its own code; the near miss beside it passes.
    #[test]
    fn each_deny_shape_is_refused_and_its_near_miss_passes() {
        let cases: [(&str, &[&str], &[&str]); 7] = [
            (
                "deny_force_push",
                &[
                    "git push --force",
                    "git push -f origin x",
                    "git push origin main --force",
                    "GIT PUSH --FORCE",
                    "git push --force-with-lease",
                    "use --force-with-lease=main:abc",
                    "`git push --force`.",
                ],
                &[
                    "git push origin x",
                    "push the branch",
                    "cargo install --force",
                    "git push --forced-name",
                    "git push origin -fix",
                    "use -f for rm only\npush later",
                ],
            ),
            (
                "deny_reset_hard",
                &[
                    "git reset --hard",
                    "git reset --hard HEAD~1",
                    "Git Reset HEAD --HARD",
                ],
                &[
                    "git reset --soft HEAD~1",
                    "reset the hard limit",
                    "--hard reset",
                ],
            ),
            (
                "deny_rm_rf",
                &[
                    "rm -rf tmp",
                    "rm -fr tmp",
                    "rm -Rf tmp",
                    "rm -rfv tmp",
                    "rm -r -f tmp",
                    "sudo rm --recursive --force /x",
                    "cd x && rm -rf y",
                ],
                &[
                    "rm -r tmp",
                    "rm -f x.log",
                    "rm tmp",
                    "perform -rf",
                    "the rm -r step\nthen -f",
                ],
            ),
            (
                "deny_drop_table",
                &[
                    "drop table users",
                    "DROP TABLE users;",
                    "drop database app",
                    "truncate table audit_log",
                    "`DROP TABLE` users",
                ],
                &[
                    "drop the table",
                    "drop tables",
                    "truncate the log",
                    "table drop",
                    "Drop the trailing comma",
                ],
            ),
            (
                "deny_visibility_public",
                &[
                    "gh repo edit --visibility public",
                    "gh repo edit --visibility=public",
                    "set visibility=public",
                    "--VISIBILITY PUBLIC",
                ],
                &[
                    "gh repo edit --visibility private",
                    "public visibility",
                    "visibility = private",
                ],
            ),
            (
                "deny_chmod_777",
                &["chmod 777 x", "chmod -R 777 dir", "chmod 0777 x"],
                &["chmod 755 x", "chmod 7777 x", "777 files"],
            ),
            (
                "deny_currency",
                &[
                    "$5",
                    "costs $20",
                    "€5",
                    "£5",
                    "₩5000",
                    "¥5",
                    "5 USD",
                    "5USD",
                    "20 eur",
                    "1,000 KRW",
                    "100원",
                    "1.5 usd.",
                    "US$5",
                ],
                &[
                    "$HOME",
                    "$PATH",
                    "${x}",
                    "5 users",
                    "USD",
                    "$ 5",
                    "5 euros",
                    "eurostar",
                    "v2 eurozone",
                    "원",
                    "rule 5",
                ],
            ),
        ];
        for (code_want, refused, passes) in cases {
            assert!(DENY.iter().any(|(c, _)| *c == code_want), "{code_want}");
            for text in refused {
                assert_eq!(deny(text), Some(code_want), "{text}");
            }
            for text in passes {
                assert_eq!(deny(text), None, "{text}");
            }
        }
    }

    /// The four ordinary pins and the text of the default request hold no shape.
    #[test]
    fn ordinary_text_holds_no_deny_shape() {
        for text in [
            "Remove the unused import?",
            "Drop the trailing comma in the fixture?",
            "Delete the dead code path?",
            "Rotate the log files daily or weekly?",
            "Which test runner does the new module use?",
            "CI runs nextest; the local docs name both runners.",
            "echo $HOME and $USER",
        ] {
            assert_eq!(deny(text), None, "{text}");
        }
    }

    /// A shape refuses in every text the gate reads: `what`, an option value, an option
    /// description, `undo` and an evidence value (any of them). The reason names the place
    /// and the shape, never the text.
    #[test]
    fn a_deny_shape_refuses_in_every_field() {
        let text = "then git push --force";
        let refusal = |edit: &dyn Fn(&mut Value)| {
            let request = parse(&request_json(edit)).unwrap();
            check(&request).unwrap_err()
        };
        for (place, edit) in [
            (
                "`what`",
                &(|v: &mut Value| v["what"] = json!(text)) as &dyn Fn(&mut Value),
            ),
            ("an option value", &|v: &mut Value| {
                v["options"][1]["value"] = json!(text)
            }),
            ("an option description", &|v: &mut Value| {
                v["options"][1]["description"] = json!(text)
            }),
            ("`undo`", &|v: &mut Value| v["undo"] = json!(text)),
            ("an evidence value", &|v: &mut Value| {
                v["evidence"] = json!({"a": "plain", "b": text})
            }),
        ] {
            let r = refusal(edit);
            assert_eq!(r.code, "deny_force_push", "{place}");
            assert!(r.reason.starts_with(place), "{}", r.reason);
            assert!(!r.reason.contains("then git"), "{}", r.reason);
        }
        // An evidence field's name is not read (it is a short identifier).
        assert_eq!(
            code(|v| v["evidence"] = json!({"rm_rf": "plain"})),
            "cleared"
        );
    }

    /// V1 audit N1: a force push written as a cluster of short flags with `f` in it, or as a
    /// refspec with a leading `+`. The near misses beside them pass.
    #[test]
    fn force_push_flag_clusters_and_plus_refspecs_are_refused() {
        for text in [
            "git push -fu origin main",
            "git push -uf origin main",
            "git push -vf",
            "git push origin +main",
            "git push origin +HEAD:main",
            "git push origin +refs/heads/x:refs/heads/x",
            "GIT PUSH ORIGIN +MAIN",
        ] {
            assert_eq!(deny(text), Some("deny_force_push"), "{text}");
        }
        for text in [
            "git push -u origin main",
            "git push origin -fix",
            "git push origin main+1",
            "push +1 to the vote",
            "git push origin main\n+main is a ref",
            "git add -f x",
        ] {
            assert_eq!(deny(text), None, "{text}");
        }
    }

    /// V1 audit N2: `gh repo create … --public` makes a repository public too.
    #[test]
    fn creating_a_public_repository_is_refused() {
        for text in [
            "gh repo create moai-x --public",
            "gh repo create --public moai-x --source .",
            "GH REPO CREATE x --PUBLIC",
        ] {
            assert_eq!(deny(text), Some("deny_visibility_public"), "{text}");
        }
        for text in [
            "gh repo create moai-x --private",
            "a --public flag elsewhere",
            "gh repo create x\n--public later",
            "gh repo view --public",
        ] {
            assert_eq!(deny(text), None, "{text}");
        }
    }

    /// V1 audit N3: the deny list reads the text after the normaliser, so a character without
    /// width, a full-width form or an ideographic space does not hide a shape; a bidi or
    /// control character is refused with its own code.
    #[test]
    fn the_deny_list_reads_normalised_text() {
        let wide = |text: &str| -> String {
            text.chars()
                .map(|c| match c {
                    ' ' => '\u{3000}',
                    c => char::from_u32(c as u32 + 0xfee0).unwrap(),
                })
                .collect()
        };
        for (text, want) in [
            ("git pu\u{200b}sh --force".to_string(), "deny_force_push"),
            ("git push --fo\u{ad}rce".to_string(), "deny_force_push"),
            (wide("git push --force"), "deny_force_push"),
            ("rm\u{3000}-rf x".to_string(), "deny_rm_rf"),
            ("git reset --\u{2060}hard".to_string(), "deny_reset_hard"),
            ("drop\u{feff} table users".to_string(), "deny_drop_table"),
            (wide("$5"), "deny_currency"),
            ("git push \u{202e}--force".to_string(), "bidi_character"),
            ("plain \u{2066}text".to_string(), "bidi_character"),
            ("plain \u{1b}text".to_string(), "control_character"),
        ] {
            assert_eq!(deny(&text), Some(want), "{text}");
        }
        // Through the gate: the place is named, the text is not.
        let request = parse(&request_json(|v| {
            v["options"][1]["description"] = json!("see \u{202e}txt.exe")
        }))
        .unwrap();
        let r = check(&request).unwrap_err();
        assert_eq!(r.code, "bidi_character");
        assert!(r.reason.starts_with("an option description"), "{}", r.reason);
        assert!(!r.reason.contains("txt"), "{}", r.reason);
    }
}
