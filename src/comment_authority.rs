//! Who may write a control line (Issue #3052).
//!
//! A control line is a line in an Issue or PR comment that a command acts on: a claim, a release
//! or handoff, an audit reservation or verdict, a close table, a gate marker. A public repository
//! lets accounts outside the organization comment, so a command honours such a line only when the
//! comment's `author_association` is `OWNER`, `MEMBER` or `COLLABORATOR`. Everything else counts
//! as "no member authority": any other value, an empty value, and a payload that lacks the field.
//! The check fails closed.
//!
//! This is the single place that decides. Every reader of comments (REST `author_association`,
//! GraphQL `authorAssociation`) puts the value into its comment type and calls [`retain_members`]
//! right where the raw JSON first becomes that type. What was dropped is returned as [`Dropped`]
//! and reported in the command's JSON as `ignored_control_lines`, so a person can see why a line
//! did not count.

use serde_json::{json, Value};

/// The reason written next to every ignored line.
pub const REASON: &str = "no_member_authority";

/// Most ignored lines listed in one answer; `count` still holds the full number.
pub const MAX_ITEMS: usize = 50;

/// `true` only for the three values that mean "a member of the organization wrote this".
///
/// The match is exact: no trimming, no case folding. A lower-case `member` (what `gh issue view`
/// prints) is an unknown value and is refused.
pub fn is_member(association: &str) -> bool {
    matches!(association, "OWNER" | "MEMBER" | "COLLABORATOR")
}

/// The `author_association` of one REST comment object; `""` when the field is missing or not a string.
pub fn rest_association(comment: &Value) -> String {
    string_field(comment, "author_association")
}

/// The `authorAssociation` of one GraphQL comment node; `""` when the field is missing or not a string.
pub fn graphql_association(comment: &Value) -> String {
    string_field(comment, "authorAssociation")
}

fn string_field(comment: &Value, key: &str) -> String {
    comment
        .get(key)
        .and_then(Value::as_str)
        .unwrap_or("")
        .to_string()
}

/// A comment type that can be checked for who wrote it.
pub trait Authored {
    fn comment_id(&self) -> i64;
    fn body(&self) -> &str;
    /// The author's association as read; `""` when the payload did not carry it.
    fn association(&self) -> &str;
}

/// What one pass removed.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Dropped {
    /// Comments that carry a control line, in the order read.
    pub lines: Vec<IgnoredLine>,
    /// Non-member comments that carry no control line (counted, not listed).
    pub other: usize,
    /// How many dropped comments had no association value at all.
    pub unread: usize,
}

/// One ignored control line.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct IgnoredLine {
    pub comment_id: i64,
    pub kind: &'static str,
    pub association: String,
}

impl Dropped {
    /// Adds what another pass dropped (a command that reads an Issue and a PR). A control line
    /// already listed under the same comment id is not listed twice.
    pub fn merge(&mut self, other: Dropped) {
        for line in other.lines {
            if !self.lines.iter().any(|l| l.comment_id == line.comment_id) {
                self.lines.push(line);
            }
        }
        self.other += other.other;
        self.unread += other.unread;
    }

    /// `true` when nothing was dropped.
    pub fn is_empty(&self) -> bool {
        self.lines.is_empty() && self.other == 0
    }

    /// The `ignored_control_lines` value for a command's JSON result.
    pub fn to_json(&self) -> Value {
        let items: Vec<Value> = self
            .lines
            .iter()
            .take(MAX_ITEMS)
            .map(|l| {
                json!({
                    "comment_id": l.comment_id,
                    "kind": l.kind,
                    "author_association": l.association,
                })
            })
            .collect();
        let mut out = json!({
            "reason": REASON,
            "count": self.lines.len(),
            "items": items,
            "other_non_member_comments": self.other,
        });
        if self.unread > 0 {
            out["hint"] = json!(
                "some comments carried no author_association value, so they were not honoured; \
                 read comments with author_association (REST) or authorAssociation (GraphQL)"
            );
        }
        out
    }
}

/// Keeps the comments a member wrote and returns what it removed.
///
/// A removed comment that carries a control line is listed with its id and kind. A removed
/// comment without one is only counted. Applying this twice is harmless.
pub fn retain_members<T: Authored>(comments: &mut Vec<T>) -> Dropped {
    let mut dropped = Dropped::default();
    comments.retain(|c| keep(&mut dropped, c.comment_id(), c.body(), c.association()));
    dropped
}

/// [`retain_members`] for the comment objects of a REST list (`id`, `body`, `author_association`),
/// applied where the raw JSON first becomes a comment type. An element that is not an object has
/// no association and is dropped.
pub fn retain_rest(comments: &mut Vec<Value>) -> Dropped {
    retain_json(comments, "id", "author_association")
}

/// [`retain_members`] for the comment nodes of a GraphQL connection (`databaseId`, `body`,
/// `authorAssociation`). The query must request `authorAssociation`; a node without it is dropped.
pub fn retain_graphql(comments: &mut Vec<Value>) -> Dropped {
    retain_json(comments, "databaseId", "authorAssociation")
}

fn retain_json(comments: &mut Vec<Value>, id_key: &str, association_key: &str) -> Dropped {
    let mut dropped = Dropped::default();
    comments.retain(|c| {
        keep(
            &mut dropped,
            c.get(id_key).and_then(Value::as_i64).unwrap_or(0),
            c.get("body").and_then(Value::as_str).unwrap_or(""),
            c.get(association_key).and_then(Value::as_str).unwrap_or(""),
        )
    });
    dropped
}

/// `true` when the comment stays; otherwise records why it went.
fn keep(dropped: &mut Dropped, id: i64, body: &str, association: &str) -> bool {
    if is_member(association) {
        return true;
    }
    if association.is_empty() {
        dropped.unread += 1;
    }
    match control_kind(body) {
        Some(kind) => dropped.lines.push(IgnoredLine {
            comment_id: id,
            kind,
            association: association.to_string(),
        }),
        None => dropped.other += 1,
    }
    false
}

/// The kind of the first control line in a body, for the report only. The decision to drop never
/// depends on it, so a kind missing here is still dropped; it is just not named.
pub fn control_kind(body: &str) -> Option<&'static str> {
    const KINDS: &[(&str, &str)] = &[
        ("moai-claim-handoff:", "claim_handoff"),
        ("moai-claim-release:", "claim_release"),
        ("moai-claim:", "claim"),
        ("moai-audit-claim", "audit_claim"),
        ("moai-resume:", "resume"),
        ("moai-tip-repair:", "tip_repair"),
        ("AUDIT #", "audit_verdict"),
        ("CLOSE-TABLE #", "close_table"),
        ("local-check:", "local_check"),
        ("ci:", "ci_line"),
        ("LANE-", "lane_line"),
        ("<!-- moai:", "marker"),
    ];
    body.lines().map(str::trim).find_map(|line| {
        KINDS
            .iter()
            .find(|(prefix, _)| line.starts_with(prefix))
            .map(|(_, kind)| *kind)
    })
}

#[cfg(test)]
mod tests {
    use super::*;

    struct C {
        id: i64,
        body: String,
        assoc: String,
    }

    impl Authored for C {
        fn comment_id(&self) -> i64 {
            self.id
        }
        fn body(&self) -> &str {
            &self.body
        }
        fn association(&self) -> &str {
            &self.assoc
        }
    }

    fn c(id: i64, assoc: &str, body: &str) -> C {
        C {
            id,
            body: body.to_string(),
            assoc: assoc.to_string(),
        }
    }

    #[test]
    fn only_owner_member_and_collaborator_are_members() {
        for ok in ["OWNER", "MEMBER", "COLLABORATOR"] {
            assert!(is_member(ok), "{ok}");
        }
        for no in [
            "NONE",
            "CONTRIBUTOR",
            "FIRST_TIMER",
            "FIRST_TIME_CONTRIBUTOR",
            "MANNEQUIN",
            "",
            " MEMBER",
            "MEMBER ",
            "member",
            "Member",
            "MEMBERS",
            "OWNER\n",
            "UNKNOWN",
        ] {
            assert!(!is_member(no), "{no:?}");
        }
    }

    #[test]
    fn a_missing_or_wrong_typed_field_reads_as_empty() {
        assert_eq!(
            rest_association(&json!({"author_association": "MEMBER"})),
            "MEMBER"
        );
        assert_eq!(rest_association(&json!({})), "");
        assert_eq!(rest_association(&json!({"author_association": null})), "");
        assert_eq!(rest_association(&json!({"author_association": 7})), "");
        assert_eq!(
            rest_association(&json!({"authorAssociation": "MEMBER"})),
            ""
        );
        assert_eq!(
            graphql_association(&json!({"authorAssociation": "OWNER"})),
            "OWNER"
        );
        assert_eq!(
            graphql_association(&json!({"author_association": "OWNER"})),
            ""
        );
        assert_eq!(graphql_association(&json!(null)), "");
    }

    #[test]
    fn members_stay_and_the_rest_go_with_their_ids() {
        let mut v = vec![
            c(
                1,
                "MEMBER",
                "moai-claim: session=a epoch=1 lease_until=2026-01-01T00:00:00Z",
            ),
            c(
                2,
                "NONE",
                "moai-claim: session=b epoch=1 lease_until=2026-01-01T00:00:00Z",
            ),
            c(3, "CONTRIBUTOR", "hello"),
            c(4, "", "AUDIT #9 head x base y — verdict: PASS"),
            c(5, "COLLABORATOR", "plain"),
        ];
        let d = retain_members(&mut v);
        assert_eq!(v.iter().map(|x| x.id).collect::<Vec<_>>(), vec![1, 5]);
        assert_eq!(
            d.lines,
            vec![
                IgnoredLine {
                    comment_id: 2,
                    kind: "claim",
                    association: "NONE".into()
                },
                IgnoredLine {
                    comment_id: 4,
                    kind: "audit_verdict",
                    association: String::new()
                },
            ]
        );
        assert_eq!(d.other, 1);
        assert_eq!(d.unread, 1);
        assert!(!d.is_empty());
    }

    #[test]
    fn a_member_only_list_drops_nothing_and_a_second_pass_is_harmless() {
        let mut v = vec![c(1, "OWNER", "moai-claim: x"), c(2, "MEMBER", "y")];
        assert!(retain_members(&mut v).is_empty());
        assert_eq!(v.len(), 2);
        let mut w = vec![c(1, "NONE", "moai-claim: x"), c(2, "MEMBER", "y")];
        retain_members(&mut w);
        let again = retain_members(&mut w);
        assert!(again.is_empty());
        assert_eq!(w.len(), 1);
    }

    #[test]
    fn the_kind_is_read_from_a_line_start_only() {
        assert_eq!(
            control_kind("intro\n  moai-claim-handoff: comment=1 to=a"),
            Some("claim_handoff")
        );
        assert_eq!(
            control_kind("moai-claim-release: comment=1"),
            Some("claim_release")
        );
        assert_eq!(
            control_kind("moai-audit-claim: lane=a head=x"),
            Some("audit_claim")
        );
        assert_eq!(
            control_kind("moai-audit-claim 철회: a (x)"),
            Some("audit_claim")
        );
        assert_eq!(control_kind("<!-- moai:gate-loop x -->"), Some("marker"));
        assert_eq!(control_kind("CLOSE-TABLE #3"), Some("close_table"));
        assert_eq!(control_kind("LANE-REPORT agent-1"), Some("lane_line"));
        assert_eq!(control_kind("see moai-claim: inside a sentence"), None);
        assert_eq!(control_kind(""), None);
    }

    #[test]
    fn the_report_caps_the_list_but_not_the_count() {
        let mut v: Vec<C> = (0..(MAX_ITEMS as i64 + 1))
            .map(|i| c(i, "NONE", "moai-claim: x"))
            .collect();
        let d = retain_members(&mut v);
        let j = d.to_json();
        assert_eq!(j["count"], MAX_ITEMS + 1);
        assert_eq!(j["items"].as_array().unwrap().len(), MAX_ITEMS);
        let mut few: Vec<C> = (0..(MAX_ITEMS as i64))
            .map(|i| c(i, "NONE", "moai-claim: x"))
            .collect();
        let j2 = retain_members(&mut few).to_json();
        assert_eq!(j2["count"], MAX_ITEMS);
        assert_eq!(j2["items"].as_array().unwrap().len(), MAX_ITEMS);
    }

    #[test]
    fn the_report_names_reason_id_kind_and_association() {
        let mut v = vec![
            c(77, "CONTRIBUTOR", "moai-claim-release: comment=1"),
            c(78, "NONE", "chat"),
        ];
        let j = retain_members(&mut v).to_json();
        assert_eq!(j["reason"], "no_member_authority");
        assert_eq!(j["count"], 1);
        assert_eq!(j["items"][0]["comment_id"], 77);
        assert_eq!(j["items"][0]["kind"], "claim_release");
        assert_eq!(j["items"][0]["author_association"], "CONTRIBUTOR");
        assert_eq!(j["other_non_member_comments"], 1);
        assert!(j.get("hint").is_none());
    }

    #[test]
    fn a_hint_appears_only_when_a_value_was_missing() {
        let mut none = vec![c(1, "NONE", "moai-claim: x")];
        assert!(retain_members(&mut none).to_json().get("hint").is_none());
        let mut empty = vec![c(1, "", "moai-claim: x")];
        let j = retain_members(&mut empty).to_json();
        assert!(j["hint"].as_str().unwrap().contains("author_association"));
    }

    #[test]
    fn a_rest_list_keeps_members_and_names_the_dropped_ids() {
        let mut v = vec![
            json!({"id": 1, "body": "moai-claim: x", "author_association": "MEMBER"}),
            json!({"id": 2, "body": "moai-claim: x", "author_association": "NONE"}),
            json!({"id": 3, "body": "AUDIT #4 x", "author_association": "CONTRIBUTOR"}),
            json!({"id": 4, "body": "moai-claim: x"}),
            json!({"id": 5, "body": "chat", "author_association": null}),
            json!({"id": 6, "body": "chat", "author_association": "OWNER"}),
            json!(null),
        ];
        let d = retain_rest(&mut v);
        let kept: Vec<i64> = v.iter().map(|c| c["id"].as_i64().unwrap()).collect();
        assert_eq!(kept, vec![1, 6]);
        let ids: Vec<i64> = d.lines.iter().map(|l| l.comment_id).collect();
        assert_eq!(ids, vec![2, 3, 4]);
        assert_eq!(d.other, 2);
        assert_eq!(d.unread, 3);
    }

    #[test]
    fn a_graphql_list_reads_database_id_and_author_association() {
        let mut v = vec![
            json!({"databaseId": 10, "body": "moai-claim: x", "authorAssociation": "COLLABORATOR"}),
            json!({"databaseId": 11, "body": "moai-claim: x", "authorAssociation": "NONE"}),
            json!({"databaseId": 12, "body": "moai-claim: x"}),
            json!({"databaseId": 13, "body": "moai-claim: x", "author_association": "MEMBER"}),
        ];
        let d = retain_graphql(&mut v);
        assert_eq!(v.len(), 1);
        assert_eq!(v[0]["databaseId"], 10);
        let ids: Vec<i64> = d.lines.iter().map(|l| l.comment_id).collect();
        assert_eq!(ids, vec![11, 12, 13]);
        assert_eq!(d.unread, 2);
    }

    #[test]
    fn two_passes_merge_into_one_report() {
        let mut a = vec![c(1, "NONE", "moai-claim: x")];
        let mut b = vec![c(2, "", "chat"), c(3, "NONE", "AUDIT #1 x")];
        let mut d = retain_members(&mut a);
        d.merge(retain_members(&mut b));
        assert_eq!(d.lines.len(), 2);
        assert_eq!(d.other, 1);
        assert_eq!(d.unread, 1);
    }

    #[test]
    fn the_same_comment_read_twice_is_listed_once() {
        let mut a = vec![c(5, "NONE", "moai-claim: x")];
        let mut b = vec![c(5, "NONE", "moai-claim: x"), c(6, "NONE", "AUDIT #1 x")];
        let mut d = retain_members(&mut a);
        d.merge(retain_members(&mut b));
        let ids: Vec<i64> = d.lines.iter().map(|l| l.comment_id).collect();
        assert_eq!(ids, vec![5, 6]);
        assert_eq!(d.to_json()["count"], 2);
    }

    /// Issue #3052 (AC4): every source file that reads the comments of an Issue or a pull
    /// request goes through this module, so a command added later cannot honour a control line
    /// that a non-member wrote. A file that touches the comments URL but reads no comment as
    /// input is listed here with its reason; a listed file that stops touching it is an error,
    /// so the list cannot go stale.
    #[test]
    fn every_source_file_that_reads_comments_filters_them_by_author() {
        // What a comment read looks like in source: the REST list and single-comment URLs, and
        // the GraphQL comment connection. Built from parts so this file does not match itself.
        let reads: Vec<String> = vec![
            format!("{}?per_page", "/comments"),
            format!("{}?", "/comments"),
            format!("{}\"", "/comments"),
            format!("{}(last:", "comments"),
            format!("{}/comments/", "issues"),
        ];
        const EXEMPT: [(&str, &str); 2] = [
            (
                "forge/mod.rs",
                "reads the Issue's own metadata and the comment URL of a write, not a comment's text",
            ),
            (
                "gate_merge.rs",
                "only posts the AUTO-DELTA and merge records; the snapshot it judges comes from gate_collect",
            ),
        ];
        let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR")).join("src");
        let mut stack = vec![root.clone()];
        let mut touching = Vec::new();
        let mut missing = Vec::new();
        while let Some(d) = stack.pop() {
            for e in std::fs::read_dir(&d).unwrap() {
                let p = e.unwrap().path();
                if p.is_dir() {
                    stack.push(p);
                    continue;
                }
                if p.extension().is_none_or(|x| x != "rs") {
                    continue;
                }
                let rel = p
                    .strip_prefix(&root)
                    .unwrap()
                    .to_string_lossy()
                    .replace('\\', "/");
                if rel == "comment_authority.rs" {
                    continue;
                }
                let text = std::fs::read_to_string(&p).unwrap();
                // Production code only: the part before the first test module.
                let prod = text.split("#[cfg(test)]").next().unwrap_or("");
                let reads_comments = prod.lines().any(|l| {
                    !l.trim_start().starts_with("//")
                        && reads.iter().any(|r| l.contains(r.as_str()))
                });
                if !reads_comments {
                    continue;
                }
                touching.push(rel.clone());
                let filters = prod.contains("comment_authority::");
                let exempt = EXEMPT.iter().any(|(f, _)| *f == rel);
                if !filters && !exempt {
                    missing.push(rel);
                }
            }
        }
        touching.sort();
        assert!(
            missing.is_empty(),
            "these files read comments without comment_authority (keep only a member's comment, \
             or add the file to EXEMPT with the reason): {missing:?}"
        );
        for (file, why) in EXEMPT {
            assert!(
                touching.iter().any(|t| t == file),
                "{file} is exempt ({why}) but no longer touches the comments URL: drop it from EXEMPT"
            );
        }
        // The scan sees what it should: a few files that must be in it.
        for want in ["claim.rs", "lane_next.rs", "gate_collect.rs", "message.rs"] {
            assert!(
                touching.iter().any(|t| t == want),
                "the scan lost {want}: {touching:?}"
            );
        }
    }
}
