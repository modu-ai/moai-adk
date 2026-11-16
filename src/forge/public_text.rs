//! Making card text safe to post on a public issue (#3071 H4, audit of #3119 B1).
//!
//! A card's title and criteria come from an issue anyone could write, and a summary is posted
//! with the user's own token. So the text is cleaned before it is shown in the plan and before it
//! is posted, in one pass that is part of the plan: what the plan shows is the cleaned text, and
//! the plan digest covers it. Five things are done, each counted so the plan can say so:
//!
//! 1. control and bidirectional-override characters (other than a line break) become a space;
//! 2. `<!--` becomes `&lt;!--`, so a title cannot carry a hidden marker such as
//!    `<!-- moai-card-summary C-9 -->` that would stop a later summary from being posted;
//! 3. `@` followed by a name gets a zero-width space after it, so nobody is notified;
//! 4. words with the shape of a credential become `[redacted]` ([`crate::redact::redact_outgoing`]);
//! 5. a line that could be read as a control line ([`crate::comment_authority::looks_like_control_line`])
//!    gets `· ` in front of it, so no command of this repository acts on it.
//!
//! This reads nothing and writes nothing; it is a pure function of the text.
use crate::comment_authority::looks_like_control_line;
use crate::redact::redact_outgoing;

/// Zero-width space, put after an `@` to stop it from being a mention.
const ZERO_WIDTH_SPACE: char = '\u{200B}';

/// What the cleaning changed, by kind.
#[derive(Clone, Debug, Default, PartialEq, Eq)]
pub struct Changes {
    pub control_characters: usize,
    pub html_comment_openers: usize,
    pub mentions: usize,
    pub secret_shapes: usize,
    pub control_lines: usize,
}

impl Changes {
    pub fn is_empty(&self) -> bool {
        *self == Changes::default()
    }

    /// One sentence per kind that changed, for the plan's `warnings`.
    pub fn warnings(&self) -> Vec<String> {
        let mut out = Vec::new();
        if self.secret_shapes > 0 {
            out.push(format!(
                "{} word(s) with the shape of a credential were replaced by [redacted]; check the source issue",
                self.secret_shapes
            ));
        }
        if self.control_lines > 0 {
            out.push(format!(
                "{} line(s) looked like a control line and got a leading `· `",
                self.control_lines
            ));
        }
        if self.mentions > 0 {
            out.push(format!(
                "{} @mention(s) were defused with a zero-width space",
                self.mentions
            ));
        }
        if self.html_comment_openers > 0 {
            out.push(format!(
                "{} HTML comment opener(s) were written as &lt;!--",
                self.html_comment_openers
            ));
        }
        if self.control_characters > 0 {
            out.push(format!(
                "{} control or direction-override character(s) were replaced by a space",
                self.control_characters
            ));
        }
        out
    }
}

/// `text` made safe to post, and what was changed.
pub fn clean_public_text(text: &str) -> (String, Changes) {
    let mut changes = Changes::default();

    let spaced: String = text
        .chars()
        .map(|c| {
            let bidi = matches!(c, '\u{202A}'..='\u{202E}' | '\u{2066}'..='\u{2069}');
            if (c.is_control() && c != '\n') || bidi {
                changes.control_characters += 1;
                ' '
            } else {
                c
            }
        })
        .collect();

    let (redacted, secrets) = redact_outgoing(&spaced);
    changes.secret_shapes = secrets;

    changes.html_comment_openers = redacted.matches("<!--").count();
    let no_comments = redacted.replace("<!--", "&lt;!--");

    let mut mentioned = String::with_capacity(no_comments.len());
    let mut chars = no_comments.chars().peekable();
    while let Some(c) = chars.next() {
        mentioned.push(c);
        if c == '@'
            && chars
                .peek()
                .is_some_and(|n| n.is_alphanumeric() || *n == '_' || *n == '-')
        {
            mentioned.push(ZERO_WIDTH_SPACE);
            changes.mentions += 1;
        }
    }

    let lines: Vec<String> = mentioned
        .split('\n')
        .map(|line| {
            if looks_like_control_line(line) {
                changes.control_lines += 1;
                format!("· {line}")
            } else {
                line.to_string()
            }
        })
        .collect();
    (lines.join("\n"), changes)
}
