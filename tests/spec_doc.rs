//! SDD-1 (SPEC-20261117-spec-tooling): the `spec.md` parser and checker, `moai_cli::spec_doc`.
//! Texts are built in memory; the bounded-read cases write into a temporary folder only.

use moai_cli::spec_doc::{
    self, check_title, id_form, new_id, parse, read_spec, render, valid_date, valid_slug, Form,
    IdForm, NewSpec, Options, Severity, SpecDoc, MAX_SPEC_BYTES, MAX_TITLE_CHARS,
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
                    accept the login.\n\n";
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
        "approval:\n  by: user\n  at: 2026-11-17T10:00:00Z\n  basis: in-session selection\n  \
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
        assert!(d.findings.is_empty(), "{minimal}: {:?}", d.findings);
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
    let bad_at = approval.replace("2026-11-17T10:00:00Z", "yesterday");
    let d = doc(&format!("{}{body}", front("approved", &bad_at)));
    assert_eq!(errors(&d), vec!["approval_field"]);
    let flow = doc(&format!(
        "{}{body}",
        front("draft", "approval: {by: user}\n")
    ));
    assert!(errors(&flow).contains(&"fm_shape"), "{:?}", flow.findings);
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
    let not_gears = REQS.replace("shall", "will");
    let d = doc(&format!("{fm}{INTENT}{not_gears}{ACS}{SCOPE}{PLAN}"));
    assert_eq!(errors(&d), vec!["req_gears"]);
    assert!(d.digest.is_none());
    // Korean body with English GEARS keywords parses.
    let ko = "## Requirements\n\n- REQ-1: When 사용자가 맞는 비밀번호를 보내면, 서버는 shall 로그인을 받아들인다.\n\n";
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
    assert_eq!(kinds.scope.write, vec!["src/a.rs", "src/b.rs", "notes/d.md"]);
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
