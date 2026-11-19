//! SDD-1 (SPEC-20261117-spec-tooling): the `spec.md` parser and checker, `moai_cli::spec_doc`.
//! Texts are built in memory; the bounded-read cases write into a temporary folder only.

use moai_cli::spec_doc::{
    self, check_title, id_form, new_id, parse, read_spec, render, valid_date, valid_slug, Form,
    IdForm, NewSpec, Options, ReqKind, Severity, SpecDoc, MAX_SPEC_BYTES, MAX_TITLE_CHARS,
};

const ID: &str = "SPEC-20261117-login-fix";

fn front(status: &str, extra: &str) -> String {
    format!(
        "---\nid: {ID}\ntitle: \"Fix login\"\nstatus: {status}\nrevision: 1\ncreated: 2026-11-17\n\
         {extra}---\n\n# Fix login\n\n"
    )
}

const INTENT: &str = "## Intent\n\nLogins fail for valid passwords.\n\n";
const REQS: &str =
    "## Requirements\n\n- REQ-1: When a user sends a valid password, the server shall \
                    accept the login.\n- REQ-2 (REQ-1): IF the password check fails, THEN the \
                    server shall refuse the login.\n\n";
const ACS: &str = "## Acceptance\n\n- AC-1 (REQ-1): A valid password logs in · verify: \
                   `cargo test --test login`\n\n";
const SCOPE: &str = "## Scope\n\n- write: `src/login.rs`\n\n";
const PLAN: &str = "## Plan\n\n1. Fix the comparison.\n\n";

fn minimal_body() -> String {
    format!("{INTENT}{REQS}{ACS}{SCOPE}{PLAN}")
}

fn doc(text: &str) -> SpecDoc {
    parse(text, &Options::default())
}

fn codes(d: &SpecDoc, s: Severity) -> Vec<&'static str> {
    d.findings
        .iter()
        .filter(|f| f.severity == s)
        .map(|f| f.code)
        .collect()
}

fn errors(d: &SpecDoc) -> Vec<&'static str> {
    codes(d, Severity::Error)
}

fn warnings(d: &SpecDoc) -> Vec<&'static str> {
    codes(d, Severity::Warning)
}

fn approval_for(body: &str) -> String {
    let draft = doc(&format!("{}{body}", front("draft", "")));
    let digest = draft.digest.expect("a GEARS draft has a digest");
    format!(
        "approval:\n  by: user\n  at: 2026-11-17\n  basis: in-session selection\n  \
         digest: {digest}\n"
    )
}

#[test]
fn a_minimal_spec_is_clean_and_minimal() {
    let d = doc(&format!("{}{}", front("draft", ""), minimal_body()));
    assert!(d.findings.is_empty(), "{:?}", d.findings);
    assert_eq!(d.form, Form::Minimal);
    assert_eq!(d.id.as_deref(), Some(ID));
    assert_eq!(d.id_form, Some(IdForm::New));
    assert_eq!(d.revision, Some(1));
    assert_eq!(d.requirements[0].id, "REQ-1");
    assert_eq!(d.acceptance[0].refs, vec!["REQ-1".to_string()]);
    assert_eq!(
        d.acceptance[0].verify.as_deref(),
        Some("cargo test --test login")
    );
    assert_eq!(d.acceptance[0].text, "A valid password logs in");
    assert_eq!(d.scope.write, vec!["src/login.rs".to_string()]);
    assert!(d.digest.is_some());
}

#[test]
fn both_templates_render_clean() {
    for minimal in [true, false] {
        let text = render(&NewSpec {
            id: ID,
            title: "Fix login",
            created: "2026-11-17",
            parent: Some("SPEC-20261101-auth"),
            minimal,
        })
        .unwrap();
        let d = parse(
            &text,
            &Options {
                folder: Some(ID.into()),
                closure: false,
            },
        );
        // A fresh draft only warns about the template text still to fill.
        assert!(errors(&d).is_empty(), "{minimal}: {:?}", d.findings);
        let w = warnings(&d);
        assert!(
            !w.is_empty() && w.iter().all(|c| *c == "placeholder"),
            "{w:?}"
        );
        assert_eq!(d.status.as_deref(), Some("draft"));
        assert_eq!(d.revision, Some(1));
        assert_eq!(d.parent.as_deref(), Some("SPEC-20261101-auth"));
        let want = if minimal { Form::Minimal } else { Form::Full };
        assert_eq!(d.form, want);
    }
}

#[test]
fn render_refuses_what_it_would_never_write() {
    let base = NewSpec {
        id: ID,
        title: "Fix login",
        created: "2026-11-17",
        parent: None,
        minimal: true,
    };
    assert!(render(&NewSpec {
        id: "SPEC-AUTH-001",
        ..base
    })
    .is_err());
    assert!(render(&NewSpec {
        title: "a \"quote\"",
        ..base
    })
    .is_err());
    assert!(render(&NewSpec {
        parent: Some(ID),
        ..base
    })
    .is_err());
    assert!(render(&NewSpec {
        parent: Some("nope"),
        ..base
    })
    .is_err());
    assert!(render(&NewSpec {
        created: "2026-02-29",
        ..base
    })
    .is_err());
    // A title that spells a placeholder stays as typed.
    let text = render(&NewSpec {
        title: "{{id}} literally",
        ..base
    })
    .unwrap();
    assert!(text.contains("title: \"{{id}} literally\""), "{text}");
}

#[test]
fn ids_new_and_legacy_and_neither() {
    let slug48 = "a".repeat(48);
    let slug49 = "a".repeat(49);
    for (id, want) in [
        ("SPEC-20261117-login-fix", Some(IdForm::New)),
        ("SPEC-20240229-leap", Some(IdForm::New)),
        ("SPEC-20261117-a1-b2-c3", Some(IdForm::New)),
        (&*format!("SPEC-20261117-{slug48}"), Some(IdForm::New)),
        (&*format!("SPEC-20261117-{slug49}"), None),
        ("SPEC-20260229-leap", None),
        ("SPEC-20261301-month", None),
        ("SPEC-20261100-day", None),
        ("SPEC-20261117-Login", None),
        ("SPEC-20261117-login--fix", None),
        ("SPEC-20261117-login-", None),
        ("SPEC-20261117-", None),
        ("SPEC-2026111-short", None),
        ("SPEC-AUTH-001", Some(IdForm::Legacy)),
        ("SPEC-V3R3-DESIGN-FOLDER-FIX-001", Some(IdForm::Legacy)),
        ("SPEC-auth-001", None),
        ("SPEC-AUTH-01", None),
        ("SPEC-AUTH-0001", None),
        ("SPEC-001", None),
        ("SPEC-3D-001", None),
        ("spec-20261117-x", None),
        ("", None),
    ] {
        assert_eq!(id_form(id), want, "{id:?}");
    }
    assert!(valid_slug(&slug48) && !valid_slug(&slug49) && !valid_slug(""));
    assert_eq!(
        new_id("2026-11-17", "login-fix").unwrap(),
        "SPEC-20261117-login-fix"
    );
    assert!(new_id("2026-11-31", "x").is_err());
    assert!(new_id("2026-11-17", "Bad").is_err());
    assert!(valid_date("2000-02-29") && !valid_date("1900-02-29") && !valid_date("2026-1-17"));
}

#[test]
fn the_id_must_match_the_folder_and_its_form() {
    let text = format!("{}{}", front("draft", ""), minimal_body());
    let other = parse(
        &text,
        &Options {
            folder: Some("SPEC-20261117-other".into()),
            closure: false,
        },
    );
    assert_eq!(errors(&other), vec!["id_folder"]);
    let bad_date = text.replace(ID, "SPEC-20261131-login-fix");
    assert_eq!(errors(&doc(&bad_date)), vec!["id_date"]);
    let bad_form = text.replace(ID, "SPEC-login-fix");
    assert_eq!(errors(&doc(&bad_form)), vec!["id_form"]);
    let none = text.replace(&format!("id: {ID}\n"), "");
    assert_eq!(errors(&doc(&none)), vec!["id_missing"]);
}

#[test]
fn every_status_word_reads_and_the_approved_ones_need_approval() {
    let body = minimal_body();
    let approval = approval_for(&body);
    for status in spec_doc::STATUSES {
        let needs = matches!(status, "approved" | "completed" | "superseded");
        let bare = doc(&format!("{}{body}", front(status, "")));
        let want: Vec<&str> = if needs {
            vec!["approval_missing"]
        } else {
            vec![]
        };
        assert_eq!(errors(&bare), want, "{status}");
        let with = doc(&format!("{}{body}", front(status, &approval)));
        assert!(errors(&with).is_empty(), "{status}: {:?}", with.findings);
        assert_eq!(with.status_as.as_deref(), Some(status));
    }
    let unknown = doc(&format!("{}{body}", front("finished", "")));
    assert_eq!(errors(&unknown), vec!["status_unknown"]);
    let missing = doc(&format!("{}{body}", front("draft", "")).replace("status: draft\n", ""));
    assert_eq!(errors(&missing), vec!["status_missing"]);
}

#[test]
fn a_go_status_maps_read_only_and_is_an_error_on_a_new_id() {
    let d = doc(&format!("{}{}", front("planned", ""), minimal_body()));
    assert_eq!(errors(&d), vec!["status_legacy"]);
    assert!(d.status_legacy);
    assert_eq!(d.status_as.as_deref(), Some("draft"));
    assert_eq!(spec_doc::legacy_status("in-progress"), Some("approved"));
    assert_eq!(spec_doc::legacy_status("draft"), None);
}

#[test]
fn a_go_generation_spec_is_read_with_warnings_only() {
    let text =
        "---\nid: SPEC-AUTH-001\nversion: \"1.0.0\"\nstatus: in-progress\ncreated: 2025-01-02\n\
                author: someone\ntags:\n  - auth\n---\n\n# Auth\n\n## Overview\n\nOld text.\n";
    let d = parse(
        text,
        &Options {
            folder: Some("SPEC-AUTH-001".into()),
            closure: false,
        },
    );
    assert!(errors(&d).is_empty(), "{:?}", d.findings);
    let w = warnings(&d);
    for code in [
        "id_legacy",
        "status_legacy",
        "fm_unknown_key",
        "title_missing",
        "revision_invalid",
        "section_unknown",
        "section_missing",
    ] {
        assert!(w.contains(&code), "{code}: {w:?}");
    }
    assert_eq!(d.status_as.as_deref(), Some("approved"));
    // The folder rule and an unknown status stay errors there.
    let moved = parse(
        text,
        &Options {
            folder: Some("SPEC-AUTH-002".into()),
            closure: false,
        },
    );
    assert_eq!(errors(&moved), vec!["id_folder"]);
    let odd = doc(&text.replace("in-progress", "whatever"));
    assert_eq!(errors(&odd), vec!["status_unknown"]);
}

#[test]
fn approval_fields_and_digest_are_checked() {
    let body = minimal_body();
    let approval = approval_for(&body);
    let ok = doc(&format!("{}{body}", front("approved", &approval)));
    assert!(ok.findings.is_empty(), "{:?}", ok.findings);
    assert_eq!(
        ok.approval.as_ref().unwrap().basis.as_deref(),
        Some("in-session selection")
    );
    // A body change after approval no longer matches the recorded digest.
    let changed = body.replace("logs in", "logs in quickly");
    let stale = doc(&format!("{}{changed}", front("approved", &approval)));
    assert_eq!(errors(&stale), vec!["approval_digest"]);
    // Prose outside REQ/AC/Scope is not in the digest.
    let prose = body.replace("Fix the comparison.", "Fix the comparison first.");
    assert!(errors(&doc(&format!("{}{prose}", front("approved", &approval)))).is_empty());
    for line in ["  by: user\n", "  basis: in-session selection\n"] {
        let partial = approval.replace(line, "");
        let d = doc(&format!("{}{body}", front("approved", &partial)));
        assert_eq!(errors(&d), vec!["approval_field"], "{line}");
    }
    // `at` is exactly YYYY-MM-DD: a time after it or another word is refused.
    for at in [
        "yesterday",
        "2026-11-17T10:00:00Z",
        "2026-11-17 10:00",
        "2026-11-31",
    ] {
        let bad_at = approval.replace("at: 2026-11-17\n", &format!("at: {at}\n"));
        let d = doc(&format!("{}{body}", front("approved", &bad_at)));
        assert_eq!(errors(&d), vec!["approval_field"], "{at:?}");
    }
    // A flow map is read like the indented block (B2).
    let flow = doc(&format!(
        "{}{body}",
        front("draft", "approval: {by: user}\n")
    ));
    assert_eq!(
        errors(&flow),
        vec!["approval_field", "approval_field", "approval_field"],
        "{:?}",
        flow.findings
    );
}

#[test]
fn frontmatter_keys_shape_and_values() {
    let body = minimal_body();
    let unknown = doc(&format!("{}{body}", front("draft", "owner: someone\n")));
    assert!(errors(&unknown).is_empty());
    assert_eq!(warnings(&unknown), vec!["fm_unknown_key"]);
    assert_eq!(unknown.unknown_keys, vec!["owner".to_string()]);
    let cards = doc(&format!("{}{body}", front("draft", "cards: C-12\n")));
    assert_eq!(errors(&cards), vec!["fm_cards"]);
    let dup = doc(&format!("{}{body}", front("draft", "status: draft\n")));
    assert_eq!(errors(&dup), vec!["fm_duplicate"]);
    let line = doc(&format!("{}{body}", front("draft", "just words\n")));
    assert_eq!(errors(&line), vec!["fm_line"]);
    let nested = doc(&format!("{}{body}", front("draft", "parent:\n  - x\n")));
    assert!(
        errors(&nested).contains(&"fm_shape"),
        "{:?}",
        nested.findings
    );
    for (from, to, code) in [
        ("revision: 1", "revision: 0", "revision_invalid"),
        ("revision: 1", "revision: 01", "revision_invalid"),
        ("revision: 1", "revision: one", "revision_invalid"),
        (
            "created: 2026-11-17",
            "created: 2026-02-30",
            "created_invalid",
        ),
        ("title: \"Fix login\"", "title: \"\"", "title_missing"),
    ] {
        let d = doc(&format!("{}{body}", front("draft", "")).replacen(from, to, 1));
        assert_eq!(errors(&d), vec![code], "{to}");
    }
    let parent = doc(&format!("{}{body}", front("draft", "parent: SPEC-x\n")));
    assert_eq!(errors(&parent), vec!["parent_invalid"]);
    let own = doc(&format!(
        "{}{body}",
        front("draft", &format!("parent: {ID}\n"))
    ));
    assert_eq!(errors(&own), vec!["parent_self"]);
    let legacy_parent = doc(&format!(
        "{}{body}",
        front("draft", "parent: SPEC-AUTH-001\n")
    ));
    assert!(errors(&legacy_parent).is_empty());
    assert_eq!(errors(&doc(&minimal_body())), vec!["fm_missing"]);
    let unclosed = doc("---\nid: x\n");
    assert_eq!(errors(&unclosed), vec!["fm_unclosed"]);
}

#[test]
fn the_title_limit_counts_characters_not_bytes() {
    let body = minimal_body();
    let n = "가".repeat(MAX_TITLE_CHARS);
    let n1 = "가".repeat(MAX_TITLE_CHARS + 1);
    assert!(n.len() > MAX_TITLE_CHARS * 2, "Hangul is three bytes");
    let at = |t: &str| {
        doc(&format!("{}{body}", front("draft", "")).replacen(
            "\"Fix login\"",
            &format!("\"{t}\""),
            1,
        ))
    };
    assert!(errors(&at(&n)).is_empty());
    assert_eq!(errors(&at(&n1)), vec!["title_long"]);
    assert!(check_title(&n).is_ok());
    assert!(check_title(&n1).is_err());
    assert!(check_title("").is_err());
    assert!(check_title("two\nlines").is_err());
}

#[test]
fn sections_minimal_full_unknown_duplicate_and_fenced() {
    let fm = front("draft", "");
    let no_intent = doc(&format!("{fm}{REQS}{ACS}{SCOPE}{PLAN}"));
    assert_eq!(errors(&no_intent), vec!["section_missing"]);
    let no_plan = doc(&format!("{fm}{INTENT}{REQS}{ACS}{SCOPE}"));
    assert_eq!(errors(&no_plan), vec!["section_missing"]);
    // Non-goals makes it full, so Risks is required too.
    let half = doc(&format!(
        "{fm}{}## Non-goals\n\n- Not the signup page.\n",
        minimal_body()
    ));
    assert_eq!(half.form, Form::Full);
    assert_eq!(errors(&half), vec!["section_missing"]);
    let full = doc(&format!(
        "{fm}{}## Non-goals\n\n- Not signup.\n\n## Risks\n\n- None known.\n\n## Decisions\n\nNone.\n",
        minimal_body()
    ));
    assert!(full.findings.is_empty(), "{:?}", full.findings);
    let empty_risks = doc(&format!(
        "{fm}{}## Non-goals\n\n- Not signup.\n\n## Risks\n\n",
        minimal_body()
    ));
    assert_eq!(errors(&empty_risks), vec!["section_empty"]);
    let extra = doc(&format!("{fm}{}## Notes\n\nx\n", minimal_body()));
    assert!(errors(&extra).is_empty());
    assert_eq!(warnings(&extra), vec!["section_unknown"]);
    let twice = doc(&format!("{fm}{}{INTENT}", minimal_body()));
    assert_eq!(errors(&twice), vec!["section_duplicate"]);
    // A heading in fenced code is text, and so is `###`.
    let fenced = doc(&format!(
        "{fm}{}```\n## Intent\n```\n\n### Intent\n",
        minimal_body()
    ));
    assert!(fenced.findings.is_empty(), "{:?}", fenced.findings);
}

#[test]
fn requirements_and_acceptance_rules() {
    let fm = front("draft", "");
    let dup_req = REQS.replace(
        "login.\n\n",
        "login.\n- REQ-1: When a user sends a bad password, the server shall refuse it.\n\n",
    );
    let d = doc(&format!("{fm}{INTENT}{dup_req}{ACS}{SCOPE}{PLAN}"));
    assert_eq!(errors(&d), vec!["req_duplicate"]);
    let not_gears = REQS.replacen("shall", "will", 1);
    let d = doc(&format!("{fm}{INTENT}{not_gears}{ACS}{SCOPE}{PLAN}"));
    assert_eq!(errors(&d), vec!["req_gears"]);
    assert!(d.digest.is_none());
    // Korean body with English GEARS and IF/THEN keywords parses.
    let ko = "## Requirements\n\n- REQ-1: When 사용자가 맞는 비밀번호를 보내면, 서버는 shall 로그인을 받아들인다.\n\
              - REQ-2 (REQ-1): IF 비밀번호가 틀리면, THEN 서버는 shall 로그인을 거절한다.\n\n";
    let d = doc(&format!("{fm}{INTENT}{ko}{ACS}{SCOPE}{PLAN}"));
    assert!(d.findings.is_empty(), "{:?}", d.findings);
    let personal = REQS.replace("valid password", "password from /Users/me/x");
    let d = doc(&format!("{fm}{INTENT}{personal}{ACS}{SCOPE}{PLAN}"));
    assert_eq!(errors(&d), vec!["req_gears"]);
    let no_req = doc(&format!(
        "{fm}{INTENT}## Requirements\n\nNone yet.\n\n{ACS}{SCOPE}{PLAN}"
    ));
    assert!(
        errors(&no_req).contains(&"req_missing"),
        "{:?}",
        no_req.findings
    );

    let acs = |text: &str| {
        doc(&format!(
            "{fm}{INTENT}{REQS}## Acceptance\n\n{text}\n{SCOPE}{PLAN}"
        ))
    };
    let next_line = acs("- AC-1: Logs in.\n  verify: cargo test\n- AC-2: Looks right.\n  manual: a person reads the page\n");
    assert!(next_line.findings.is_empty(), "{:?}", next_line.findings);
    assert_eq!(
        next_line.acceptance[0].verify.as_deref(),
        Some("cargo test")
    );
    assert_eq!(
        next_line.acceptance[1].manual.as_deref(),
        Some("a person reads the page")
    );
    let unknown = acs("- AC-1: Fast enough · unknown: no benchmark yet\n");
    assert!(errors(&unknown).is_empty());
    assert_eq!(warnings(&unknown), vec!["ac_unknown"]);
    assert_eq!(errors(&acs("- AC-1: Logs in.\n")), vec!["ac_no_verify"]);
    assert_eq!(
        errors(&acs("- AC-1: Logs in · verify:\n")),
        vec!["ac_no_verify"]
    );
    // A `verify:` inside the text, not after a space or `·`, is not a tag.
    assert_eq!(errors(&acs("- AC-1: reverify:x\n")), vec!["ac_no_verify"]);
    // A line that does not start with the tag does not verify the AC above it.
    assert_eq!(
        errors(&acs("- AC-1: Logs in.\n  see verify: later\n")),
        vec!["ac_no_verify"]
    );
    assert_eq!(
        errors(&acs("- AC-1: a · verify: x\n- AC-1: b · verify: y\n")),
        vec!["ac_duplicate"]
    );
    assert_eq!(
        errors(&acs("- AC-1 (REQ-1, REQ-7): a · verify: x\n")),
        vec!["ac_ref_unknown"]
    );
    assert!(errors(&acs("Nothing here.\n")).contains(&"ac_missing"));
}

/// The EARS unwanted-behaviour form (`IF …, THEN … shall …`) and the failure-path rule: every
/// `When` requirement is named by an IF/THEN requirement or says `failure: none`.
#[test]
fn when_requirements_need_a_failure_path() {
    let fm = front("draft", "");
    let reqs = |lines: &str| {
        doc(&format!(
            "{fm}{INTENT}## Requirements\n\n{lines}\n{ACS}{SCOPE}{PLAN}"
        ))
    };
    let when1 = "- REQ-1: When a user logs in, the server shall check the password.\n";
    let when3 = "- REQ-3: When a user logs out, the server shall end the session.\n";
    let if1 = "- REQ-2 (REQ-1): IF the check fails, THEN the server shall refuse the login.\n";
    let if3 = "- REQ-4 (REQ-3): IF no session exists, THEN the server shall answer 404.\n";

    // Zero covered: the When requirement alone is an error at its line.
    let alone = reqs(when1);
    assert_eq!(errors(&alone), vec!["req_failure_path"]);
    assert_eq!(alone.findings[0].line, Some(17));
    // N covered: clean, and the IF requirement reads as unwanted.
    let covered = reqs(&format!("{when1}{if1}"));
    assert!(covered.findings.is_empty(), "{:?}", covered.findings);
    assert_eq!(covered.requirements[1].kind, ReqKind::Unwanted);
    assert_eq!(covered.requirements[1].refs, vec!["REQ-1".to_string()]);
    assert_eq!(covered.requirements[0].kind, ReqKind::Gears);
    // N+1: a second When requirement without its own IF is the one error.
    let one_more = reqs(&format!("{when1}{if1}{when3}"));
    assert_eq!(errors(&one_more), vec!["req_failure_path"]);
    assert!(
        one_more.findings[0].message.contains("REQ-3"),
        "{:?}",
        one_more.findings
    );
    assert!(reqs(&format!("{when1}{if1}{when3}{if3}"))
        .findings
        .is_empty());
    // An IF requirement that names no REQ covers nothing.
    let unnamed = reqs(&format!(
        "{when1}- REQ-2: IF the check fails, THEN the server shall refuse the login.\n"
    ));
    assert_eq!(errors(&unnamed), vec!["req_failure_path"]);
    // Naming a REQ that is not there.
    let ghost = reqs(&format!(
        "{when1}- REQ-2 (REQ-9): IF the check fails, THEN the server shall refuse it.\n"
    ));
    assert_eq!(errors(&ghost), vec!["req_ref_unknown", "req_failure_path"]);
    // The trailing `· failure: none` note (with an optional ` — reason`) is the explicit
    // exemption; it is not part of the text (N5).
    for line in [
        "- REQ-1: When a user logs in, the server shall log the time · failure: none\n",
        "- REQ-1: When a user logs in, the server shall log the time. · failure: none — a log line only\n",
    ] {
        let d = reqs(line);
        assert!(d.findings.is_empty(), "{line}: {:?}", d.findings);
        assert!(d.requirements[0].failure_none, "{line}");
        assert!(!d.requirements[0].text.contains("failure"), "{line}");
    }
    // Any other spelling or place is not the note.
    for line in [
        "- REQ-1: When a user logs in, the server shall log the time. failure: none\n",
        "- REQ-1: When a user logs in, the server shall log the time (failure: none)\n",
        "- REQ-1: When a user logs in, the server shall log nofailure: none\n",
        "- REQ-1: When a user logs in · failure: none, the server shall log the time\n",
    ] {
        let d = reqs(line);
        assert!(!d.requirements[0].failure_none, "{line}");
        assert!(
            errors(&d).contains(&"req_failure_path"),
            "{line}: {:?}",
            d.findings
        );
    }
    // Only `When` triggers the rule.
    let states = reqs(
        "- REQ-1: While the cache is warm, the server shall answer from it.\n\
         - REQ-2: Where audit is on, the server shall record each login.\n\
         - REQ-3: The server shall hash every password.\n",
    );
    assert!(states.findings.is_empty(), "{:?}", states.findings);
    // IF/THEN shapes: case, a missing comma, and what is refused.
    let lower = reqs(&format!(
        "{when1}- REQ-2 (REQ-1): if the check fails then the server shall refuse it.\n"
    ));
    assert!(lower.findings.is_empty(), "{:?}", lower.findings);
    for bad in [
        "- REQ-2 (REQ-1): IF , THEN the server shall refuse it.\n",
        "- REQ-2 (REQ-1): IF the check fails, the server shall refuse it.\n",
        "- REQ-2 (REQ-1): IF the check fails, THEN the server refuses it.\n",
        "- REQ-2 (REQ-1): IF the check fails, THEN shall refuse it.\n",
        "- REQ-2 (REQ-1): IF the token is ghp_abc, THEN the server shall refuse it.\n",
    ] {
        let d = reqs(&format!("{when1}{bad}"));
        assert_eq!(errors(&d), vec!["req_gears", "req_failure_path"], "{bad}");
        assert!(d.digest.is_none(), "{bad}");
    }
    // A Go-generation SPEC gets the rule as a warning.
    let legacy = doc(
        &format!("{fm}{INTENT}## Requirements\n\n{when1}\n{ACS}{SCOPE}{PLAN}")
            .replace(ID, "SPEC-AUTH-001"),
    );
    assert!(errors(&legacy).is_empty(), "{:?}", legacy.findings);
    assert!(warnings(&legacy).contains(&"req_failure_path"));
}

/// B2: the YAML shapes a SPEC is written with — a flow map for `approval` and `null`.
#[test]
fn frontmatter_reads_flow_maps_and_null() {
    let body = minimal_body();
    let block = approval_for(&body);
    let digest = block.split("digest: ").nth(1).unwrap().trim().to_string();
    let flow = format!(
        "parent: null\napproval: {{by: user, at: 2026-11-17, basis: \"in-session: yes, all\", \
         digest: {digest}}}\n"
    );
    let d = doc(&format!("{}{body}", front("approved", &flow)));
    assert!(d.findings.is_empty(), "{:?}", d.findings);
    let a = d.approval.unwrap();
    assert_eq!(a.basis.as_deref(), Some("in-session: yes, all"));
    assert_eq!(a.digest.as_deref(), Some(digest.as_str()));
    assert_eq!(d.parent, None);
    // `null` reads as absent: an approved SPEC without a digest is refused.
    let null_digest = flow.replace(&digest, "null");
    let d = doc(&format!("{}{body}", front("approved", &null_digest)));
    assert_eq!(errors(&d), vec!["approval_field"]);
    // A flow map that is not closed, or one on a key that takes one value.
    let open = doc(&format!(
        "{}{body}",
        front("draft", "approval: {by: user\n")
    ));
    assert!(errors(&open).contains(&"fm_shape"), "{:?}", open.findings);
    let title = doc(&format!("{}{body}", front("draft", "")).replacen(
        "title: \"Fix login\"",
        "title: {a: b}",
        1,
    ));
    assert_eq!(errors(&title), vec!["fm_shape"]);
}

/// B2: `(REQ-1…REQ-3)` and `(REQ-1..REQ-3)` name every requirement in the range.
#[test]
fn a_range_names_at_most_a_thousand_ids() {
    let fm = front("draft", "");
    let refs = |r: &str| {
        let d = doc(&format!(
            "{fm}{INTENT}{REQS}## Acceptance\n\n- AC-1 ({r}): all · verify: x\n\n{SCOPE}{PLAN}"
        ));
        d.acceptance[0].refs.clone()
    };
    // N: 1000 ids are read; N+1: 1001 ids stay one literal ref (unknown).
    assert_eq!(refs("REQ-1..REQ-1000").len(), 1000);
    assert_eq!(refs("REQ-1..REQ-1001"), vec!["REQ-1..REQ-1001".to_string()]);
}

#[test]
fn refs_take_ranges() {
    let fm = front("draft", "");
    let three = "## Requirements\n\n- REQ-1: The server shall a.\n- REQ-2: The server shall b.\n\
                 - REQ-3: The server shall c.\n\n";
    let ac = |refs: &str| {
        doc(&format!(
            "{fm}{INTENT}{three}## Acceptance\n\n- AC-1 ({refs}): all · verify: x\n\n{SCOPE}{PLAN}"
        ))
    };
    for refs in [
        "REQ-1…REQ-3",
        "REQ-1..REQ-3",
        "REQ-1 … REQ-3",
        "REQ-2, REQ-1..REQ-1",
    ] {
        let d = ac(refs);
        assert!(d.findings.is_empty(), "{refs}: {:?}", d.findings);
    }
    assert_eq!(
        ac("REQ-1…REQ-3").acceptance[0].refs,
        vec!["REQ-1", "REQ-2", "REQ-3"]
    );
    for refs in ["REQ-3..REQ-1", "REQ-1..REQ-4", "전체", "REQ-1..AC-2"] {
        assert!(
            errors(&ac(refs)).contains(&"ac_ref_unknown"),
            "{refs}: {:?}",
            ac(refs).findings
        );
    }
}

/// B3: `[NEEDS CLARIFICATION…]` and an unfilled template placeholder are errors in an approved
/// or completed SPEC and warnings in a draft; text in fenced code and other `<…>` are free.
#[test]
fn placeholders_and_clarifications_follow_the_status() {
    let body = minimal_body();
    let cases = [
        (
            "Logins fail for valid passwords.",
            "Logins fail. [NEEDS CLARIFICATION: which?]",
            "needs_clarification",
        ),
        ("Fix the comparison.", "<first step>", "placeholder"),
        (
            "Fix the comparison.",
            "Fix <the step REQ-1 needs fails> here",
            "placeholder",
        ),
    ];
    for (from, to, code) in cases {
        let changed = body.replace(from, to);
        let approval = approval_for(&changed);
        for (status, error) in [
            ("draft", false),
            ("approved", true),
            ("completed", true),
            ("rejected", false),
        ] {
            let extra = if status == "draft" || status == "rejected" {
                ""
            } else {
                &approval
            };
            let d = doc(&format!("{}{changed}", front(status, extra)));
            let (e, w) = (errors(&d), warnings(&d));
            if error {
                assert_eq!((e, w), (vec![code], vec![]), "{status} {to}");
            } else {
                assert_eq!((e, w), (vec![], vec![code]), "{status} {to}");
            }
        }
    }
    // Fenced code and `<…>` that is not template text are free.
    let free = body.replace(
        "Fix the comparison.",
        "Read `<state>/specs/` and <n>.\n\n```\n[NEEDS CLARIFICATION]\n<first step>\n```",
    );
    let d = doc(&format!("{}{free}", front("draft", "")));
    assert!(d.findings.is_empty(), "{:?}", d.findings);
}

/// B3: the size budget warns above 12 requirements (N=12 passes, 13 warns) and above 400
/// lines (400 passes, 401 warns); it is never an error.
#[test]
fn the_size_budget_warns_above_twelve_requirements_and_four_hundred_lines() {
    let fm = front("draft", "");
    let reqs = |n: usize| {
        let lines: String = (1..=n)
            .map(|i| format!("- REQ-{i}: The server shall do step {i}.\n"))
            .collect();
        doc(&format!(
            "{fm}{INTENT}## Requirements\n\n{lines}\n## Acceptance\n\n- AC-1 (REQ-1): ok · verify: x\n\n{SCOPE}{PLAN}"
        ))
    };
    assert!(reqs(12).findings.is_empty(), "{:?}", reqs(12).findings);
    assert_eq!(warnings(&reqs(13)), vec!["size_requirements"]);
    assert!(errors(&reqs(13)).is_empty());
    let base = format!("{fm}{}", minimal_body());
    let used = base.lines().count();
    let pad = |total: usize| {
        let extra = "Plan detail.\n".repeat(total - used);
        doc(&format!("{base}{extra}"))
    };
    assert_eq!(
        format!("{base}{}", "x\n".repeat(400 - used))
            .lines()
            .count(),
        400
    );
    assert!(pad(400).findings.is_empty(), "{:?}", pad(400).findings);
    assert_eq!(warnings(&pad(401)), vec!["size_lines"]);
    assert!(errors(&pad(401)).is_empty());
}

/// N2: the secret-shape and personal-path checks of `parse_gears` hold for AC and Scope text.
#[test]
fn ac_and_scope_text_get_the_text_checks() {
    let fm = front("draft", "");
    let with_ac = |ac: &str| {
        doc(&format!(
            "{fm}{INTENT}{REQS}## Acceptance\n\n- AC-1 (REQ-1): {ac}\n\n{SCOPE}{PLAN}"
        ))
    };
    for ac in [
        "Logs in from /Users/me/x · verify: cargo test",
        "Logs in · verify: curl -H token=abc",
        "Logs in · manual: read ~/notes",
        "Logs in · verify: echo ghp_abcdef",
    ] {
        assert_eq!(errors(&with_ac(ac)), vec!["text_unsafe"], "{ac}");
    }
    assert!(with_ac("Logs in · verify: cargo test --test login")
        .findings
        .is_empty());
    let with_scope = |s: &str| doc(&format!("{fm}{INTENT}{REQS}{ACS}## Scope\n\n{s}\n{PLAN}"));
    assert_eq!(
        errors(&with_scope(
            "- write: `src/a.rs`\n- forbidden: `password=x`\n"
        )),
        vec!["text_unsafe"]
    );
    assert!(with_scope("- write: `src/a.rs`\n").findings.is_empty());
}

/// A tag line may wrap: indented lines below it that are not list items add to its kind.
#[test]
fn a_wrapped_scope_tag_line_keeps_its_kind() {
    let fm = front("draft", "");
    let d = doc(&format!(
        "{fm}{INTENT}{REQS}{ACS}## Scope\n\n- write: `a.md`, `b.md`,\n  `src/c.rs`\n- read-only: \
         `src/d.rs`,\n  `e/**`\n- `f.rs`\n  `not-a-path-g`\n\n{PLAN}"
    ));
    assert!(d.findings.is_empty(), "{:?}", d.findings);
    assert_eq!(d.scope.write, vec!["a.md", "b.md", "src/c.rs", "f.rs"]);
    assert_eq!(d.scope.read_only, vec!["src/d.rs", "e/**"]);
}

#[test]
fn scope_paths_are_relative_and_inside() {
    let fm = front("draft", "");
    let scope = |text: &str| {
        doc(&format!(
            "{fm}{INTENT}{REQS}{ACS}## Scope\n\n{text}\n{PLAN}"
        ))
    };
    let kinds = scope(
        "- write: `src/a.rs`, `src/b.rs`\n- read-only: `src/c.rs`\n- forbidden: `.gitignore`\n- `notes/d.md` and why\n",
    );
    assert!(kinds.findings.is_empty(), "{:?}", kinds.findings);
    // A tag with paths names only its own; an untagged item is `write`.
    assert_eq!(
        kinds.scope.write,
        vec!["src/a.rs", "src/b.rs", "notes/d.md"]
    );
    assert_eq!(kinds.scope.read_only, vec!["src/c.rs"]);
    assert_eq!(kinds.scope.forbidden, vec![".gitignore"]);
    // A tag line without paths sets the kind of the items below it.
    let listed = scope("write:\n- src/x.rs\nforbidden:\n- .env\n");
    assert_eq!(listed.scope.write, vec!["src/x.rs"]);
    assert_eq!(listed.scope.forbidden, vec![".env"]);
    for (path, code) in [
        ("/etc/passwd", "scope_absolute"),
        ("~/notes", "scope_absolute"),
        ("\\\\server\\x", "scope_absolute"),
        ("C:\\x", "scope_absolute"),
        ("../other", "scope_parent"),
        ("src/../../x", "scope_parent"),
    ] {
        let d = scope(&format!("- write: `{path}`\n"));
        assert_eq!(errors(&d), vec![code], "{path}");
    }
    // `..` as part of a name is not a climb.
    assert!(scope("- write: `src/a..b.rs`\n").findings.is_empty());
    assert_eq!(errors(&scope("Nothing.\n")), vec!["scope_empty"]);
    assert_eq!(
        errors(&scope("- read-only: `src/a.rs`\n")),
        vec!["scope_empty"]
    );
}

#[test]
fn closure_needs_completed_and_every_ac_met_or_withdrawn() {
    let two = format!(
        "{INTENT}{REQS}## Acceptance\n\n- AC-1 (REQ-1): a · verify: x\n- AC-2: b · manual: look\n\n{SCOPE}{PLAN}"
    );
    let approval = approval_for(&two);
    let closed = |status: &str, result: &str| {
        parse(
            &format!("{}{two}{result}", front(status, &approval)),
            &Options {
                folder: None,
                closure: true,
            },
        )
    };
    let good = "## Result\n\n- AC-1: met — CI run 1\n- AC-2: withdrawn — out of scope\n";
    assert!(
        closed("completed", good).findings.is_empty(),
        "{:?}",
        closed("completed", good).findings
    );
    assert_eq!(errors(&closed("approved", good)), vec!["closure_status"]);
    assert_eq!(errors(&closed("completed", "")), vec!["closure_result"]);
    let open = "## Result\n\n- AC-1: met\n- AC-2: pending\n";
    assert_eq!(errors(&closed("completed", open)), vec!["closure_open"]);
    let missing = "## Result\n\n- AC-1: met\n";
    assert_eq!(errors(&closed("completed", missing)), vec!["closure_open"]);
    let extra = format!("{good}- AC-3: met\n");
    assert_eq!(
        errors(&closed("completed", &extra)),
        vec!["closure_unknown"]
    );
    // Without --closure, a Result section is just read.
    let plain = doc(&format!("{}{two}{open}", front("approved", &approval)));
    assert!(plain.findings.is_empty(), "{:?}", plain.findings);
    assert_eq!(plain.result.len(), 2);
}

#[test]
fn reads_are_bounded_regular_utf8_files() {
    let dir = tempfile::tempdir().unwrap();
    let at = |name: &str, bytes: &[u8]| {
        let p = dir.path().join(name);
        std::fs::write(&p, bytes).unwrap();
        p
    };
    let limit = MAX_SPEC_BYTES as usize;
    assert_eq!(limit, 262_144);
    let n = at("n.md", &vec![b'a'; limit]);
    assert_eq!(read_spec(&n).unwrap().len(), limit);
    let n1 = at("n1.md", &vec![b'a'; limit + 1]);
    assert_eq!(read_spec(&n1).unwrap_err().code, "read_too_large");
    assert_eq!(read_spec(&at("empty.md", b"")).unwrap(), "");
    let bad = at("bad.md", &[0xff, 0xfe]);
    assert_eq!(read_spec(&bad).unwrap_err().code, "read_utf8");
    assert_eq!(
        read_spec(&dir.path().join("none.md")).unwrap_err().code,
        "read_missing"
    );
    assert_eq!(read_spec(dir.path()).unwrap_err().code, "read_not_file");
    #[cfg(unix)]
    {
        let link = dir.path().join("link.md");
        std::os::unix::fs::symlink(&n, &link).unwrap();
        assert_eq!(read_spec(&link).unwrap_err().code, "read_not_file");
        // A SPEC folder that is a link is not entered.
        let real = dir.path().join("real");
        std::fs::create_dir(&real).unwrap();
        std::fs::write(real.join("spec.md"), b"x").unwrap();
        let folder = dir.path().join(ID);
        std::os::unix::fs::symlink(&real, &folder).unwrap();
        let d = spec_doc::load(&folder, false);
        assert_eq!(errors(&d), vec!["read_not_file"]);
    }
}

// ---- SPEC-20261119-enforce-interview-record -------------------------------------------------

const NEW_ID: &str = "SPEC-20261119-login-fix";

fn front_on(created: &str, status: &str, extra: &str) -> String {
    format!(
        "---\nid: {NEW_ID}\ntitle: \"Fix login\"\nstatus: {status}\nrevision: 1\ncreated: {created}\n\
         {extra}---\n\n# Fix login\n\n"
    )
}

/// An approved SPEC created on `created` with `interview` after its body.
fn approved_on(created: &str, status: &str, interview: &str) -> String {
    let body = minimal_body();
    let digest = doc(&format!("{}{body}", front_on(created, "draft", "")))
        .digest
        .expect("a GEARS draft has a digest");
    let approval = format!(
        "approval:\n  by: user\n  at: {created}\n  basis: in-session selection\n  \
         digest: {digest}\n"
    );
    format!("{}{body}{interview}", front_on(created, status, &approval))
}

const GOOD_INTERVIEW: &str =
    "## Interview\n\n- Q1: Which logins fail? → A: Valid passwords only.\n\
    - Q2: Is the API contract fixed? -> A: Yes, no change.\n\n";

/// REQ-1, REQ-4: an approved SPEC created on or after 2026-11-19 needs its interview; an older
/// one, a draft and a Go-generation SPEC do not.
#[test]
fn interview_is_required() {
    for status in ["approved", "completed", "superseded"] {
        for created in ["2026-11-19", "2026-11-20", "2027-01-01"] {
            let d = doc(&approved_on(created, status, ""));
            assert_eq!(errors(&d), vec!["interview_missing"], "{status} {created}");
            let ok = doc(&approved_on(created, status, GOOD_INTERVIEW));
            assert!(
                errors(&ok).is_empty() && warnings(&ok).is_empty(),
                "{status} {created}: {:?}",
                ok.findings
            );
        }
    }
    // Created the day before: not asked.
    let old = doc(&approved_on("2026-11-18", "approved", ""));
    assert!(old.findings.is_empty(), "{:?}", old.findings);
    // A draft is not asked, on any date.
    let draft = doc(&format!(
        "{}{}",
        front_on("2026-11-19", "draft", ""),
        minimal_body()
    ));
    assert!(draft.findings.is_empty(), "{:?}", draft.findings);
    // Other statuses before approval are not asked either.
    for status in ["proposed", "in-review", "planned"] {
        let d = doc(&format!(
            "{}{}",
            front_on("2026-11-19", status, ""),
            minimal_body()
        ));
        assert!(
            !d.findings.iter().any(|f| f.code.starts_with("interview")),
            "{status}: {:?}",
            d.findings
        );
    }
    // A `created` that is no calendar date has its own error and is not asked for an interview.
    let bad = doc(&approved_on("2026-13-45", "approved", ""));
    assert!(
        !bad.findings.iter().any(|f| f.code.starts_with("interview")),
        "{:?}",
        bad.findings
    );
    // A Go-generation SPEC is read, never asked.
    let legacy = approved_on("2026-11-19", "approved", "").replace(NEW_ID, "SPEC-AUTH-001");
    let d = doc(&legacy);
    assert!(
        !d.findings.iter().any(|f| f.code.starts_with("interview")),
        "{:?}",
        d.findings
    );
}

/// REQ-2, REQ-3: the two entry forms, and the six shapes that are not entries.
#[test]
fn interview_entries() {
    let none = "## Interview\n\n- none: The change is a typo in one comment.\n\n";
    let ok = doc(&approved_on("2026-11-19", "approved", none));
    assert!(ok.findings.is_empty(), "{:?}", ok.findings);
    // An arrow inside the question is part of the question: the arrow before `A:` counts.
    let arrows = "## Interview\n\n- Q1: Does a -> b hold, or a → c? -> A: a -> b holds.\n\n";
    let ok = doc(&approved_on("2026-11-19", "approved", arrows));
    assert!(ok.findings.is_empty(), "{:?}", ok.findings);
    // Prose lines between entries are read as prose.
    let prose = "## Interview\n\nAsked in the planning session.\n\n- Q1: Why? → A: Because.\n\n";
    let ok = doc(&approved_on("2026-11-19", "approved", prose));
    assert!(ok.findings.is_empty(), "{:?}", ok.findings);

    for (what, section) in [
        ("not an entry form", "## Interview\n\n- just some text\n\n"),
        ("empty question", "## Interview\n\n- Q1: → A: yes\n\n"),
        ("empty answer", "## Interview\n\n- Q1: why? → A:\n\n"),
        ("reason too short", "## Interview\n\n- none: short\n\n"),
        (
            "none beside a question",
            "## Interview\n\n- none: A reason that is long enough.\n- Q1: x → A: y\n\n",
        ),
        (
            "no entry at all",
            "## Interview\n\nNothing was written here.\n\n",
        ),
    ] {
        let approved = doc(&approved_on("2026-11-19", "approved", section));
        assert_eq!(errors(&approved), vec!["interview_entry"], "{what}");
        // The same shape in a SPEC that is not approved is a warning.
        let draft = doc(&format!(
            "{}{}{section}",
            front_on("2026-11-19", "draft", ""),
            minimal_body()
        ));
        assert!(errors(&draft).is_empty(), "{what}: {:?}", draft.findings);
        assert_eq!(warnings(&draft), vec!["interview_entry"], "{what}");
        // An approved SPEC created before the date is held to the same entry rules.
        let old = doc(&approved_on("2026-11-18", "approved", section));
        assert_eq!(errors(&old), vec!["interview_entry"], "{what} (old)");
    }
    // The reason's minimum is ten characters, not bytes.
    let nine = "## Interview\n\n- none: 아홉글자입니다아아\n\n";
    let ten = "## Interview\n\n- none: 열글자입니다열글자야\n\n";
    assert_eq!(
        errors(&doc(&approved_on("2026-11-19", "approved", nine))),
        vec!["interview_entry"]
    );
    assert!(doc(&approved_on("2026-11-19", "approved", ten))
        .findings
        .is_empty());
}

/// REQ-5: `## Interview` is a known section, once, and outside the digest.
#[test]
fn interview_section() {
    let with = doc(&format!(
        "{}{}{GOOD_INTERVIEW}",
        front_on("2026-11-19", "draft", ""),
        minimal_body()
    ));
    assert!(with.findings.is_empty(), "{:?}", with.findings);
    assert!(with.sections.iter().any(|s| s == "Interview"));
    let without = doc(&format!(
        "{}{}",
        front_on("2026-11-19", "draft", ""),
        minimal_body()
    ));
    assert_eq!(
        with.digest, without.digest,
        "the interview is not in the digest"
    );
    let twice = doc(&format!(
        "{}{}{GOOD_INTERVIEW}{GOOD_INTERVIEW}",
        front_on("2026-11-19", "draft", ""),
        minimal_body()
    ));
    assert_eq!(errors(&twice), vec!["section_duplicate"]);
    // A SPEC of the older generation without the section reads as before.
    let plain = doc(&approved_on("2026-11-17", "approved", ""));
    assert!(plain.findings.is_empty(), "{:?}", plain.findings);
}

// ---- SPEC-20261119-interview-guidance -------------------------------------------------------

fn interview_findings(d: &SpecDoc) -> Vec<&'static str> {
    d.findings
        .iter()
        .filter(|f| f.code.starts_with("interview"))
        .map(|f| f.code)
        .collect()
}

/// REQ-1, REQ-2: both templates start with an Interview section whose example entry reads as
/// a well-formed entry, and an approved SPEC that keeps the example is refused at its line.
#[test]
fn interview_template() {
    for minimal in [true, false] {
        let text = render(&NewSpec {
            id: NEW_ID,
            title: "Fix login",
            created: "2026-11-19",
            parent: None,
            minimal,
        })
        .unwrap();
        assert!(text.contains("\n## Interview\n"), "{minimal}: no section");
        let opts = Options {
            folder: Some(NEW_ID.into()),
            closure: false,
        };
        let d = parse(&text, &opts);
        assert!(
            d.sections.iter().any(|s| s == "Interview"),
            "{minimal}: {:?}",
            d.sections
        );
        assert!(
            interview_findings(&d).is_empty(),
            "{minimal}: {:?}",
            d.findings
        );
        // The example entry holds placeholders: once approved, they are an error at its line.
        let line = text
            .lines()
            .position(|l| l.trim_start().starts_with("- Q1:"))
            .expect("an example entry")
            + 1;
        let approved = parse(&text.replace("status: draft", "status: approved"), &opts);
        assert!(
            approved.findings.iter().any(|f| f.code == "placeholder"
                && f.severity == Severity::Error
                && f.line == Some(line)),
            "{minimal} line {line}: {:?}",
            approved.findings
        );
        assert!(
            interview_findings(&approved).is_empty(),
            "{minimal}: {:?}",
            approved.findings
        );
    }
}

/// REQ-3 (audit N1, N2): only approved, completed and superseded SPECs are asked, and a
/// broken entry is an error in each of them.
#[test]
fn interview_statuses() {
    for status in ["rejected", "archived", "draft"] {
        let d = doc(&approved_on("2026-11-19", status, ""));
        assert!(
            interview_findings(&d).is_empty(),
            "{status}: {:?}",
            d.findings
        );
    }
    for status in ["approved", "completed", "superseded"] {
        let d = doc(&approved_on("2026-11-19", status, ""));
        assert_eq!(
            interview_findings(&d),
            vec!["interview_missing"],
            "{status}"
        );
        assert!(errors(&d).contains(&"interview_missing"), "{status}");
        let broken = doc(&approved_on(
            "2026-11-19",
            status,
            "## Interview\n\n- just some text\n\n",
        ));
        assert_eq!(
            interview_findings(&broken),
            vec!["interview_entry"],
            "{status}"
        );
        assert!(errors(&broken).contains(&"interview_entry"), "{status}");
        assert!(!warnings(&broken).contains(&"interview_entry"), "{status}");
    }
}

/// REQ-5 (audit N3, N4): the edges this check does not close, named in the SPEC's Risks. When
/// one of them is closed, the test and the Risks change together.
#[test]
fn interview_edges() {
    let hidden = "<!--\n## Interview\n\n- Q1: x → A: y\n-->\n\n";
    let d = doc(&approved_on("2026-11-19", "approved", hidden));
    assert!(interview_findings(&d).is_empty(), "{:?}", d.findings);
    let fenced = "## Interview\n\n```\n- Q1: x → A: y\n```\n\n";
    let d = doc(&approved_on("2026-11-19", "approved", fenced));
    assert!(interview_findings(&d).is_empty(), "{:?}", d.findings);
    // A hand-written Go-form id is read, never asked.
    let go = approved_on("2026-11-19", "approved", "").replace(NEW_ID, "SPEC-AUTH-001");
    let d = doc(&go);
    assert!(interview_findings(&d).is_empty(), "{:?}", d.findings);
}
