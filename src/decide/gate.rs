//! The gate of `moai decide` (card C-37 S2): which choices a model may make at all.
//!
//! SKILL.md sorts an open choice: one that cannot be undone, that touches one of the five HOLD
//! categories or that changes a `guarded` setting goes to the user and waits. Only what is
//! left, a reversible choice of category `none`, is the leader's, and only that may be put to
//! a model in the leader's place. So the gate lets through exactly: `reversible` is `true`,
//! `category` is `none`, the default is one of the options, and the text itself does not read
//! as one of the other classes.
//!
//! The last check is [`screen`], a fixed word list. It is a deterministic filter, not a
//! classifier: it cannot show a text is harmless, it only catches a `none` that names money,
//! a key, a deletion and the like in plain words. A hit costs nothing but the model's turn
//! (the leader decides instead), so the list leans towards refusing: after the audit of the
//! first round it also holds everyday verbs (`remove`, `drop`, `buy`, `rotate`), and an
//! ordinary engineering choice that uses one (`remove the unused import`) goes to the leader
//! too (leader decision 2026-11-17, audit N8).

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

    /// For the tests of the steps after the gate, which must not depend on its word list.
    #[cfg(test)]
    pub(crate) fn unchecked(request: &'a Request) -> Self {
        Cleared { request }
    }
}

/// The classes the screen knows and the words that name them. The five HOLD categories and
/// `guarded` are the shared ones ([`CATEGORIES`]); `external_transmission` and `irreversible`
/// are the other two reasons SKILL.md gives for a choice that waits for the user. An ASCII
/// term matches whole words (a plural `s` allowed), in any letter case; a Korean term matches
/// anywhere, since particles attach to the word.
pub const SCREEN: [(&str, &[&str]); 8] = [
    (
        "money",
        &[
            "money",
            "payment",
            "billing",
            "invoice",
            "purchase",
            "subscription",
            "refund",
            "pricing",
            "paid plan",
            "budget",
            "cost",
            "wire",
            "charge",
            "buy",
            "결제",
            "과금",
            "요금",
            "구매",
            "환불",
            "비용",
            "예산",
            "돈",
            "송금",
        ],
    ),
    (
        "legal",
        &[
            "legal",
            "license",
            "licence",
            "lawsuit",
            "copyright",
            "trademark",
            "terms of service",
            "gdpr",
            "법적",
            "법률",
            "라이선스",
            "저작권",
            "약관",
        ],
    ),
    (
        "external_contract",
        &[
            "contract",
            "sla",
            "nda",
            "vendor agreement",
            "sign the",
            "계약",
        ],
    ),
    (
        "secret",
        &[
            "secret",
            "credential",
            "password",
            "passphrase",
            "api key",
            "api_key",
            "api-key",
            "private key",
            "access token",
            "auth token",
            "비밀번호",
            "비밀 키",
            "자격 증명",
            "암호",
            "시크릿",
        ],
    ),
    (
        "real_data",
        &[
            "real data",
            "production data",
            "production database",
            "customer data",
            "personal data",
            "user data",
            "pii",
            "prod",
            "production",
            "customer",
            "실제 데이터",
            "운영",
            "운영 데이터",
            "고객 데이터",
            "개인정보",
        ],
    ),
    (
        "guarded",
        &[
            "permission",
            "global setting",
            "branch protection",
            "force-push",
            "force push",
            "rm -rf",
            "reset --hard",
            "delete",
            "deletion",
            "drop",
            "wipe",
            "erase",
            "truncate",
            "remove",
            "--force",
            "two-factor",
            "2fa",
            "certificate",
            "rotate",
            "권한",
            "전역 설정",
            "보호 규칙",
            "삭제",
            "지우",
            "지울",
            "지워",
        ],
    ),
    (
        "external_transmission",
        &[
            "upload",
            "publish",
            "third party",
            "third-party",
            "external service",
            "send to",
            "make public",
            "go public",
            "private repository",
            "repository public",
            "외부 전송",
            "외부로",
            "업로드",
        ],
    ),
    (
        "irreversible",
        &[
            "irreversible",
            "cannot be undone",
            "can't be undone",
            "one-way",
            "되돌릴 수 없",
        ],
    ),
];

/// The first class `text` reads as, with the term that matched, or `None`.
pub fn screen(text: &str) -> Option<(&'static str, &'static str)> {
    let lower = text.to_lowercase();
    SCREEN.iter().find_map(|(class, terms)| {
        terms
            .iter()
            .find(|term| holds_term(&lower, term))
            .map(|term| (*class, *term))
    })
}

/// True when the lower-cased text holds `term`: anywhere for a Korean term, as whole words
/// for an ASCII one (no ASCII letter or digit right before it or right after it, a plural `s`
/// aside).
fn holds_term(lower: &str, term: &str) -> bool {
    if !term.is_ascii() {
        return lower.contains(term);
    }
    let in_word = |c: char| c.is_ascii_alphanumeric();
    lower.match_indices(term).any(|(at, _)| {
        let before = lower[..at].chars().next_back();
        let after = &lower[at + term.len()..];
        let after = after.strip_prefix('s').unwrap_or(after);
        !before.is_some_and(in_word) && !after.chars().next().is_some_and(in_word)
    })
}

/// The gate: `request` as a [`Cleared`] choice, or the first rule that refuses it. The order
/// is fixed: reversible, category, default, then the text screen.
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
    let texts = [request.what.as_str(), request.evidence.as_str()]
        .into_iter()
        .chain(
            request
                .options
                .iter()
                .flat_map(|o| [o.value.as_str(), o.description.as_str()]),
        );
    for text in texts {
        if let Some((class, term)) = screen(text) {
            return Err(Refusal::new(
                "class_mismatch",
                format!(
                    "`category` is `none`, but the text reads as `{class}` (term `{term}`): \
                     the leader decides"
                ),
            ));
        }
    }
    Ok(Cleared { request })
}

#[cfg(test)]
mod tests {
    use super::super::request::parse;
    use super::super::testing::request_json;
    use super::*;
    use serde_json::json;

    fn code(edit: impl FnOnce(&mut serde_json::Value)) -> &'static str {
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
    /// wrong with it, and a category is judged before the default.
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
    }

    #[test]
    fn the_screen_reads_each_class_in_english_and_korean() {
        for (text, class) in [
            ("Do we renew the subscription this month?", "money"),
            ("이번 달 결제를 진행할까요", "money"),
            ("Is the MIT License compatible here?", "legal"),
            ("Does the SLA allow a night deploy?", "external_contract"),
            ("Rotate the API key now or later?", "secret"),
            ("Store two passwords or one?", "secret"),
            ("Run the migration on production data first?", "real_data"),
            ("개인정보가 든 표를 먼저 옮길까요", "real_data"),
            ("Delete the old branches or keep them?", "guarded"),
            ("전역 설정을 바꿀까요", "guarded"),
            ("Upload the report or attach it?", "external_transmission"),
            ("This cannot be undone: squash now?", "irreversible"),
        ] {
            assert_eq!(screen(text).map(|(c, _)| c), Some(class), "{text}");
        }
    }

    /// Audit N8: the auditor's probe sentences, each declared `none`, each refused. The two
    /// from the first round stay refused.
    #[test]
    fn the_audit_probes_are_refused() {
        for text in [
            "Drop the users table in prod now or after the release?",
            "Wire 4,000 dollars to the contractor today or next week?",
            "Paste the production key into the CI variables or keep it local?",
            "Email the full customer list to the agency today?",
            "Charge the card 20 USD a month for the Pro tier or stay on Free?",
            "Wipe the backups and truncate the audit log, or keep them?",
            "Sign the MSA with Acme now?",
            "Run push --force to main or open a PR?",
            "Erase the user records older than 30 days?",
            "Make the private repository public today?",
            "Turn off two-factor login for the admins?",
            "Remove the prod database volume?",
            "Buy the enterprise seats now?",
            "Rotate the prod signing certificate and tokens tonight?",
            "운영 DB 를 지울까요 아니면 둘까요",
            "돈을 오늘 보낼까요 내일 보낼까요",
            "Delete the cache folder?",
            "Which paid plan do we take?",
        ] {
            assert!(screen(text).is_some(), "{text}");
            assert_eq!(
                code(|v| v["what"] = json!(text)),
                "class_mismatch",
                "{text}"
            );
            assert_eq!(
                code(|v| v["evidence"] = json!(text)),
                "class_mismatch",
                "{text}"
            );
        }
    }

    /// Each word N8 adds, alone, in the class it reads as.
    #[test]
    fn the_added_words_each_read_as_a_class() {
        for (text, class) in [
            ("wire it", "money"),
            ("charge them", "money"),
            ("buy it", "money"),
            ("돈", "money"),
            ("송금", "money"),
            ("sign the form", "external_contract"),
            ("prod", "real_data"),
            ("production", "real_data"),
            ("customer", "real_data"),
            ("운영", "real_data"),
            ("drop it", "guarded"),
            ("wipe it", "guarded"),
            ("erase it", "guarded"),
            ("truncate it", "guarded"),
            ("remove it", "guarded"),
            ("use --force", "guarded"),
            ("two-factor", "guarded"),
            ("2fa", "guarded"),
            ("certificate", "guarded"),
            ("rotate it", "guarded"),
            ("지우기", "guarded"),
            ("지울까", "guarded"),
            ("지워요", "guarded"),
            ("make public", "external_transmission"),
            ("go public", "external_transmission"),
            ("the private repository", "external_transmission"),
        ] {
            assert_eq!(screen(text).map(|(c, _)| c), Some(class), "{text}");
        }
        // Whole words still: these only contain an added word.
        for text in [
            "a product page",
            "the rewire step",
            "signing off",
            "dropdown menu",
            "buyer",
        ] {
            assert_eq!(screen(text), None, "{text}");
        }
    }

    /// The miss direction: a term inside a longer word is not the term, and ordinary
    /// engineering text passes.
    #[test]
    fn the_screen_matches_whole_words_only() {
        for text in [
            "Which test runner does the new module use?",
            // `sla` and `nda`, `cost`, `delete`, `publish` and `pii` inside longer words.
            "Translate the island agenda",
            "The costly path or the cheap one",
            "Deleted rows are filtered upstream",
            "A republished crate name",
            "Use spiing or polling",
            "nextest 와 cargo test 중 무엇을 쓸까요",
        ] {
            assert_eq!(screen(text), None, "{text}");
        }
        // Punctuation and case do not hide a term.
        assert_eq!(screen("(PASSWORD)").map(|(c, _)| c), Some("secret"));
        assert_eq!(screen("budget.").map(|(c, _)| c), Some("money"));
    }

    /// Declared `none`, read as another class: refused, in whichever text field it stands.
    /// The reason names the class and the fixed term, not the text.
    #[test]
    fn a_none_that_reads_as_another_class_is_a_mismatch() {
        for field in ["what", "evidence"] {
            let request = parse(&request_json(|v| {
                v[field] = json!("We could delete the cache folder first.")
            }))
            .unwrap();
            let r = check(&request).unwrap_err();
            assert_eq!(r.code, "class_mismatch", "{field}");
            assert!(r.reason.contains("guarded") && r.reason.contains("`delete`"));
            assert!(!r.reason.contains("cache folder"), "{}", r.reason);
        }
        assert_eq!(
            code(|v| v["options"][0]["description"] = json!("pay the invoice")),
            "class_mismatch"
        );
        assert_eq!(
            code(|v| {
                v["options"][0]["value"] = json!("purchase");
                v["default"] = json!("purchase");
            }),
            "class_mismatch"
        );
    }
}
