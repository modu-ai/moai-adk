//! The text of a `moai decide` request that would leave the machine (card C-37 S2): the
//! question, the options and the evidence, and nothing else. The card, the default and the
//! deadline are never sent.
//!
//! Two rules, in this order:
//! 1. A credential shape anywhere in that text refuses the whole request (user decision
//!    2026-11-16, D-secret-hit: refuse and fall back, zero calls; never mask and send). The
//!    shapes are the shared ones of [`crate::redact::redact_outgoing`], read a second time
//!    with punctuation taken as a word break, so a key glued to a label (`x:<key>`) counts.
//! 2. The home folder and the user name are masked ([`mask`]): they say who and where, and a
//!    decision needs neither. An option value that would change under the mask is refused
//!    instead, because the answer names an option by its value.

use super::gate::Cleared;
use super::Refusal;

/// What stands in for the user's name, and for the home folder.
pub const USER_MASK: &str = "[user]";
pub const HOME_MASK: &str = "~";

/// Whose home and name to mask. Given by the caller, so nothing here reads the environment.
#[derive(Debug, Clone, Default, PartialEq)]
pub struct Who {
    pub home: Option<String>,
    pub user: Option<String>,
}

/// The masked texts, made only by [`build`]: the one thing the provider code reads.
#[derive(Debug, Clone, PartialEq)]
pub struct Outbound {
    what: String,
    options: Vec<(String, String)>,
    evidence: String,
}

impl Outbound {
    pub fn what(&self) -> &str {
        &self.what
    }

    /// `(value, description)` in the request's order.
    pub fn options(&self) -> &[(String, String)] {
        &self.options
    }

    pub fn evidence(&self) -> &str {
        &self.evidence
    }
}

/// True when `text` holds a credential shape (see the module rules).
pub fn holds_secret_shape(text: &str) -> bool {
    let apart: String = text
        .chars()
        .map(|c| if WORD_BREAKS.contains(c) { ' ' } else { c })
        .collect();
    crate::redact::redact_outgoing(text).1 > 0 || crate::redact::redact_outgoing(&apart).1 > 0
}

/// Punctuation read as a word break by the second pass of [`holds_secret_shape`]. Not `-`,
/// `_` or `.`: they are part of the shapes themselves.
const WORD_BREAKS: &str = ":=,;|/\\\"'`()[]{}<>@?&#!";

/// Home-shaped folders: the path part after one of these is a person's name.
const HOME_PARENTS: [&str; 3] = ["/Users/", "/home/", "\\Users\\"];

/// Where a name after a [`HOME_PARENTS`] prefix ends (or at white space).
const NAME_ENDS: &str = "/\\\"'`)]}>,;:";

/// Account names that say nothing about a person and are ordinary words in a question
/// (`the repository root`): never masked as a name.
const PLAIN_NAMES: [&str; 1] = ["root"];

/// `text` with the home folder as `~`, the user name as a whole word as `[user]`, and the
/// name in any other home-shaped path (`/Users/<name>`, `/home/<name>`, `\Users\<name>`) as
/// `[user]`. The name is `who.user` and the last part of `who.home`.
pub fn mask(text: &str, who: &Who) -> String {
    let mut out = text.to_string();
    let home = who
        .home
        .as_deref()
        .map(|h| h.trim_end_matches(['/', '\\']))
        .filter(|h| h.len() > 1);
    if let Some(home) = home {
        // The folder itself, not a sibling whose name only starts the same way.
        out = replace_outside_words(&out, home, HOME_MASK, |_, after| {
            after.is_some_and(|c| c.is_alphanumeric() || c == '_' || c == '-')
        });
    }
    let from_home = home.and_then(|h| h.rsplit(['/', '\\']).next());
    for name in [who.user.as_deref(), from_home].into_iter().flatten() {
        if name.is_empty() || PLAIN_NAMES.contains(&name) {
            continue;
        }
        let in_word = |c: char| c.is_alphanumeric() || c == '_';
        out = replace_outside_words(&out, name, USER_MASK, |before, after| {
            before.is_some_and(in_word) || after.is_some_and(in_word)
        });
    }
    for parent in HOME_PARENTS {
        out = mask_names_after(&out, parent);
    }
    out
}

/// `text` with each `find` replaced by `with`, except where `inside(before, after)` says the
/// characters around it make it part of something longer.
fn replace_outside_words(
    text: &str,
    find: &str,
    with: &str,
    inside: impl Fn(Option<char>, Option<char>) -> bool,
) -> String {
    let mut out = String::with_capacity(text.len());
    let mut from = 0;
    for (at, _) in text.match_indices(find) {
        let end = at + find.len();
        if inside(text[..at].chars().next_back(), text[end..].chars().next()) {
            continue;
        }
        out.push_str(&text[from..at]);
        out.push_str(with);
        from = end;
    }
    out.push_str(&text[from..]);
    out
}

/// `text` with the path part after each `parent` replaced by `[user]`.
fn mask_names_after(text: &str, parent: &str) -> String {
    let mut out = String::with_capacity(text.len());
    let mut rest = text;
    while let Some(at) = rest.find(parent) {
        let (head, tail) = rest.split_at(at + parent.len());
        out.push_str(head);
        let end = tail
            .find(|c: char| c.is_whitespace() || NAME_ENDS.contains(c))
            .unwrap_or(tail.len());
        if end == 0 || tail.starts_with(USER_MASK) {
            rest = tail;
            continue;
        }
        out.push_str(USER_MASK);
        rest = &tail[end..];
    }
    out.push_str(rest);
    out
}

/// The outbound text of a cleared request, or the refusal of rule 1 or of a masked option
/// value.
pub fn build(cleared: &Cleared<'_>, who: &Who) -> Result<Outbound, Refusal> {
    let request = cleared.request();
    let sent = [request.what.as_str(), request.evidence.as_str()]
        .into_iter()
        .chain(
            request
                .options
                .iter()
                .flat_map(|o| [o.value.as_str(), o.description.as_str()]),
        );
    for text in sent {
        if holds_secret_shape(text) {
            return Err(Refusal::new(
                "secret_shape",
                "the text to send holds a credential shape: nothing is sent, the leader decides",
            ));
        }
    }
    if request
        .options
        .iter()
        .any(|o| mask(&o.value, who) != o.value)
    {
        return Err(Refusal::new(
            "option_value_private",
            "an option value holds the home folder or the user name: give the option a short \
             name instead",
        ));
    }
    Ok(Outbound {
        what: mask(&request.what, who),
        options: request
            .options
            .iter()
            .map(|o| (o.value.clone(), mask(&o.description, who)))
            .collect(),
        evidence: mask(&request.evidence, who),
    })
}

#[cfg(test)]
mod tests {
    use super::super::request::parse;
    use super::super::testing::{request_json, who};
    use super::*;
    use serde_json::json;

    /// Past the gate without asking it: these cases are about this step alone.
    fn built(edit: impl FnOnce(&mut serde_json::Value)) -> Result<Outbound, Refusal> {
        let request = parse(&request_json(edit)).unwrap();
        build(&Cleared::unchecked(&request), &who())
    }

    /// Test data: shaped like credentials, built at run time, none real.
    fn fakes() -> Vec<String> {
        let mut all: Vec<String> = crate::redact::secret_shapes()
            .into_iter()
            .filter(|shape| *shape != "sk-")
            .map(|shape| format!("{shape}{}", "0a".repeat(20)))
            .collect();
        all.push(format!("sk-{}", "a1".repeat(12)));
        all.push(format!("AKIA{}", "A1".repeat(8)));
        all.push(format!("xoxb-{}", "1".repeat(12)));
        all
    }

    #[test]
    fn a_secret_shape_in_any_sent_field_refuses_the_request() {
        for fake in fakes() {
            assert!(holds_secret_shape(&format!("see {fake} here")), "{fake}");
            for field in ["what", "evidence"] {
                let r =
                    built(|v| v[field] = json!(format!("it printed {fake} twice"))).unwrap_err();
                assert_eq!(r.code, "secret_shape", "{field}");
                assert!(!r.reason.contains(&fake));
            }
            let r = built(|v| v["options"][1]["description"] = json!(fake.clone())).unwrap_err();
            assert_eq!(r.code, "secret_shape");
        }
    }

    /// A key glued to a label or wrapped in punctuation is still a key.
    #[test]
    fn punctuation_does_not_hide_a_secret_shape() {
        let key = format!("sk-{}", "a1".repeat(12));
        for text in [
            format!("Authorization:{key}"),
            format!("header=Bearer,{key};"),
            format!("{{\"k\":\"{key}\"}}"),
            format!("see <{key}>"),
        ] {
            assert!(holds_secret_shape(&text), "{text}");
        }
    }

    /// The miss direction: ordinary words that only contain a shape's letters are sent.
    #[test]
    fn ordinary_text_is_not_a_secret_shape() {
        for text in [
            "a risk-based, task-level choice",
            "disk-full handling: ask-first or retry",
            "the token limit and the password rule (no value here)",
            "sk-short",
            "CI runs nextest; the local docs name both runners.",
        ] {
            assert!(!holds_secret_shape(text), "{text}");
        }
        assert!(built(|v| v["evidence"] = json!("a risk-based, task-level choice")).is_ok());
    }

    #[test]
    fn the_home_folder_and_the_user_name_are_masked() {
        let w = who();
        for (text, want) in [
            ("/Users/testperson/proj/Cargo.toml", "~/proj/Cargo.toml"),
            ("see /Users/testperson.", "see ~."),
            ("cd /Users/testperson && ls", "cd ~ && ls"),
            // Another folder that only starts like the home: its name is masked as a name.
            ("/Users/testperson2/x", "/Users/[user]/x"),
            // Someone else's home, on each system's layout.
            (
                "/Users/other/x and /home/alice/y",
                "/Users/[user]/x and /home/[user]/y",
            ),
            ("C:\\Users\\bob\\src", "C:\\Users\\[user]\\src"),
            // The name alone, as a whole word only.
            (
                "owner: testperson, branch testperson/fix",
                "owner: [user], branch [user]/fix",
            ),
            (
                "testpersons and mytestperson stay",
                "testpersons and mytestperson stay",
            ),
            ("no path here", "no path here"),
        ] {
            assert_eq!(mask(text, &w), want, "{text}");
        }
    }

    #[test]
    fn masking_uses_only_what_it_is_given() {
        // Nobody to mask: only the home-shaped paths change.
        let nobody = Who::default();
        assert_eq!(
            mask("/home/alice/x testperson", &nobody),
            "/home/[user]/x testperson"
        );
        // A home without a name still gives one: its last part.
        let home_only = Who {
            home: Some("/home/carol".into()),
            user: None,
        };
        assert_eq!(mask("/home/carol/a by carol", &home_only), "~/a by [user]");
        // A home of `/` or an empty name masks nothing by itself.
        let root = Who {
            home: Some("/".into()),
            user: Some(String::new()),
        };
        assert_eq!(mask("/etc/hosts and more", &root), "/etc/hosts and more");
        // The account `root` names nobody: its home is masked, the word stays a word.
        let superuser = Who {
            home: Some("/root".into()),
            user: Some("root".into()),
        };
        assert_eq!(
            mask("/root/proj is the repository root", &superuser),
            "~/proj is the repository root"
        );
    }

    #[test]
    fn the_sent_fields_are_masked_and_nothing_else_is_kept() {
        let out = built(|v| {
            v["what"] = json!("Which runner for /Users/testperson/proj?");
            v["options"][0]["description"] = json!("as testperson set it up");
            v["evidence"] = json!("/Users/testperson/proj/ci.yml runs nextest");
        })
        .unwrap();
        assert_eq!(out.what(), "Which runner for ~/proj?");
        assert_eq!(
            out.options(),
            [
                ("nextest".to_string(), "as [user] set it up".to_string()),
                ("cargo-test".to_string(), "plain cargo test".to_string()),
            ]
        );
        assert_eq!(out.evidence(), "~/proj/ci.yml runs nextest");
    }

    /// An option value is what the answer names, so it is never rewritten: a value the mask
    /// would change refuses the request.
    #[test]
    fn an_option_value_the_mask_would_change_is_refused() {
        let r = built(|v| {
            v["options"][0]["value"] = json!("testperson");
            v["default"] = json!("testperson");
        })
        .unwrap_err();
        assert_eq!(r.code, "option_value_private");
        assert!(!r.reason.contains("testperson"));
    }
}
