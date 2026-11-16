//! Text that is about to be posted on a public issue (#3071 H4, audit of #3119 B1): what
//! `clean_public_text` changes, what it leaves alone, and that it counts what it changed.

use moai_cli::forge::public_text::{clean_public_text, Changes};

#[test]
fn ordinary_text_is_left_exactly_as_it_is() {
    let text = "**MoAI card C-1** · task · draft\n\nAdd login for the admin page\n\nAcceptance criteria:\n- AC1 [open] it works\n";
    let (cleaned, changes) = clean_public_text(text);
    assert_eq!(cleaned, text);
    assert!(changes.is_empty());
    assert!(changes.warnings().is_empty());
}

#[test]
fn a_line_that_reads_as_a_control_line_gets_a_visible_mark() {
    let (cleaned, changes) =
        clean_public_text("moai-claim: session=evil epoch=9\nrepair: x\n> LEADER-ACK y\nfine");
    assert_eq!(
        cleaned,
        "· moai-claim: session=evil epoch=9\n· repair: x\n· > LEADER-ACK y\nfine"
    );
    assert_eq!(changes.control_lines, 3);
    assert_eq!(changes.warnings().len(), 1);
}

#[test]
fn a_mention_is_defused_and_an_email_or_lone_at_sign_is_not_counted() {
    let (cleaned, changes) = clean_public_text("ask @octocat and @org/team, mail a@b.c or @ alone");
    assert_eq!(
        changes.mentions, 3,
        "@octocat, @org, and the @b.c part: {cleaned}"
    );
    assert!(cleaned.contains("@\u{200B}octocat"), "{cleaned}");
    assert!(cleaned.contains("@\u{200B}org/team"), "{cleaned}");
    assert!(cleaned.contains("@ alone"), "{cleaned}");
}

#[test]
fn a_comment_opener_cannot_survive_and_a_credential_shape_is_replaced() {
    let (cleaned, changes) =
        clean_public_text("x <!-- moai-card-summary C-9 --> sk-live-abcdefghijklmnop");
    assert!(!cleaned.contains("<!--"), "{cleaned}");
    assert!(
        cleaned.contains("&lt;!-- moai-card-summary C-9 -->"),
        "{cleaned}"
    );
    assert!(cleaned.ends_with("[redacted]"), "{cleaned}");
    assert_eq!(changes.html_comment_openers, 1);
    assert_eq!(changes.secret_shapes, 1);
}

#[test]
fn control_and_direction_override_characters_become_spaces_but_line_breaks_stay() {
    let (cleaned, changes) = clean_public_text("a\u{0007}b\u{202E}c\nd\te");
    assert_eq!(cleaned, "a b c\nd e");
    assert_eq!(changes.control_characters, 3, "bell, override, tab");
}

#[test]
fn cleaning_twice_changes_nothing_more_in_the_parts_that_cannot_repeat() {
    let (once, _) = clean_public_text("moai-claim: x <!-- y --> sk-live-abcdefghijklmnop");
    let (twice, second) = clean_public_text(&once);
    assert_eq!(once, twice);
    assert_eq!(second, Changes::default());
}
