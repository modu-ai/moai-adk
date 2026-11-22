//! SPEC-20261122-html-report-skill AC-1 to AC-4: the moai package carries the internal
//! html-report skill (`skills/moai-html-report`). The four tests read a fixed list of the eleven
//! package files and walk no folder: the frontmatter and the withdrawn Go-era text (AC-1), the
//! html head fonts and the mustache braces of the six templates (AC-2), the artifact page text
//! (AC-3) and the delivery order of SKILL.md (AC-4).

use std::fs;
use std::path::Path;

const DIR: &str = "skills/moai-html-report";
const REFERENCES: [&str; 4] = [
    "artifact-contract.md",
    "craft-fundamentals.md",
    "design-tokens.md",
    "fonts.md",
];
const TEMPLATES: [&str; 6] = ["status", "incident", "plan", "explainer", "financial", "pr"];

const PRETENDARD: &str = r#"<link rel="stylesheet" href="https://cdn.jsdelivr.net/gh/orioncactus/pretendard@v1.3.9/dist/web/static/pretendard.css">"#;
const MONO: &str = r#"<link rel="stylesheet" href="https://fonts.googleapis.com/css2?family=JetBrains+Mono:wght@400;500&display=swap">"#;
const MERMAID: &str = "https://cdn.jsdelivr.net/npm/mermaid@11.4.1/dist/mermaid.min.js";

/// The eleven package files, relative to the skill folder.
fn skill_files() -> Vec<String> {
    let mut files = vec!["SKILL.md".to_string()];
    files.extend(REFERENCES.iter().map(|r| format!("references/{r}")));
    files.extend(template_files());
    files
}

fn template_files() -> Vec<String> {
    TEMPLATES
        .iter()
        .map(|t| format!("references/templates/{t}.html.mustache"))
        .collect()
}

fn read(rel: &str) -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(DIR).join(rel);
    fs::read_to_string(&path).unwrap_or_else(|e| panic!("{DIR}/{rel}: {e}"))
}

/// Byte offsets of `.moai/reports/` that are not the user-level `~/.moai/reports/`: a report
/// folder inside the project.
fn project_report_paths(text: &str) -> Vec<usize> {
    text.match_indices(".moai/reports/")
        .map(|(i, _)| i)
        .filter(|&i| !text[..i].ends_with("~/"))
        .collect()
}

fn is_html_field(name: &str) -> bool {
    name.ends_with("_html") || name == "diagram_svg"
}

/// Every mustache tag whose braces do not fit its field: an HTML field (`*_html`,
/// `diagram_svg`) needs three braces, every other field two. Section tags (`#`, `/`, `^`),
/// comments (`!`) and partials (`>`) are not fields; `{{&x}}` is refused (use three braces).
fn brace_errors(text: &str) -> Vec<String> {
    let mut errors = Vec::new();
    let mut rest = text;
    while let Some(start) = rest.find("{{") {
        let after = &rest[start + 2..];
        if let Some(inner) = after.strip_prefix('{') {
            let end = inner
                .find("}}}")
                .unwrap_or_else(|| panic!("unclosed three-brace tag: {inner:.20}"));
            let name = inner[..end].trim();
            if !is_html_field(name) {
                errors.push(format!("three braces on a text field: {name}"));
            }
            rest = &inner[end + 3..];
        } else {
            let end = after
                .find("}}")
                .unwrap_or_else(|| panic!("unclosed tag: {after:.20}"));
            let name = after[..end].trim();
            if name.starts_with('&') {
                errors.push(format!("`&` tag, use three braces: {name}"));
            } else if !name.starts_with(['#', '/', '^', '!', '>']) && is_html_field(name) {
                errors.push(format!("two braces on an HTML field: {name}"));
            }
            rest = &after[end + 2..];
        }
    }
    errors
}

/// The `## Delivery` section of SKILL.md, up to the next `## ` heading.
fn delivery_section(skill: &str) -> &str {
    let start = skill
        .find("\n## Delivery\n")
        .expect("SKILL.md has a `## Delivery` section");
    let body = &skill[start + 1..];
    let end = body[3..].find("\n## ").map_or(body.len(), |i| i + 3);
    &body[..end]
}

#[test]
fn the_package_carries_the_internal_skill_with_no_go_era_text() {
    // The project-folder matcher in both directions.
    assert!(project_report_paths("in `~/.moai/reports/moai-adk/`").is_empty());
    assert_eq!(project_report_paths("write `.moai/reports/x.md`"), vec![7]);
    assert_eq!(project_report_paths("`$HOME/.moai/reports/`").len(), 1);

    for rel in skill_files() {
        let path = Path::new(env!("CARGO_MANIFEST_DIR")).join(DIR).join(&rel);
        assert!(path.is_file(), "{DIR}/{rel} is missing");
    }

    let skill = read("SKILL.md");
    assert!(
        skill.starts_with("---\n"),
        "SKILL.md starts with frontmatter"
    );
    let end = skill[4..]
        .find("\n---\n")
        .expect("SKILL.md frontmatter is closed");
    let frontmatter: Vec<&str> = skill[4..4 + end].lines().collect();
    assert!(
        frontmatter.contains(&"name: moai-html-report"),
        "{frontmatter:?}"
    );
    assert!(
        frontmatter.contains(&"user-invocable: false"),
        "{frontmatter:?}"
    );

    for rel in skill_files() {
        let text = read(&rel);
        for word in [
            "report.yaml",
            "report.format",
            "Codex",
            "skill-routing doctrine",
            "moai-domain-",
        ] {
            assert!(!text.contains(word), "{rel} still names `{word}`");
        }
        assert_eq!(
            project_report_paths(&text),
            Vec::<usize>::new(),
            "{rel} names a report folder inside the project"
        );
    }
}

#[test]
fn the_html_head_loads_pretendard_and_html_fields_use_three_braces() {
    // The brace checker in both directions.
    assert!(brace_errors("{{{tldr_html}}} {{{diagram_svg}}} {{title}} {{.}}").is_empty());
    assert!(brace_errors("{{#callout_html}}x{{/callout_html}} {{^met}}y{{/met}}").is_empty());
    assert_eq!(brace_errors("<p>{{tldr_html}}</p>").len(), 1);
    assert_eq!(brace_errors("{{diagram_svg}}").len(), 1);
    assert_eq!(brace_errors("<li>{{{.}}}</li>").len(), 1);
    assert_eq!(brace_errors("{{{title}}} {{&body}}").len(), 2);

    for rel in template_files() {
        let text = read(&rel);
        let head = &text[..text
            .find("</head>")
            .unwrap_or_else(|| panic!("{rel}: no </head>"))];
        assert!(head.contains(PRETENDARD), "{rel}: no Pretendard link");
        assert!(head.contains(MONO), "{rel}: no JetBrains Mono link");
        assert!(
            head.contains(r#"<html lang="{{lang}}">"#),
            "{rel}: lang is not filled"
        );
        assert!(
            head.contains("<title>{{title}}</title>"),
            "{rel}: the title is not a name"
        );
        assert!(!text.contains("Noto"), "{rel} names a Noto font");
        assert!(
            !text.contains("artifact:begin"),
            "{rel} keeps the artifact block"
        );
        assert_eq!(brace_errors(&text), Vec::<String>::new(), "{rel}");
    }

    let skill = read("SKILL.md");
    assert!(
        skill.contains(MERMAID),
        "SKILL.md names the exact mermaid script"
    );
    assert!(
        skill.contains("mmd-fallback"),
        "SKILL.md names the visible fallback"
    );
    assert!(
        !skill.contains("mermaid@11/"),
        "SKILL.md keeps a version range"
    );
}

#[test]
fn the_artifact_text_names_google_fonts_noto_by_language_and_pre_mermaid() {
    let text = read("references/artifact-contract.md");
    for name in [
        "fonts.googleapis.com",
        "fonts.gstatic.com",
        "Noto Sans KR",
        "Noto Sans JP",
        "Noto Sans SC",
        "Noto Sans TC",
    ] {
        assert!(
            text.contains(name),
            "artifact-contract.md does not name {name}"
        );
    }
    assert!(
        text.contains(r#"pre class="mermaid""#) || text.contains("pre.mermaid"),
        "artifact-contract.md does not name pre.mermaid"
    );
    assert!(
        !text.contains("cdn.jsdelivr.net/gh/orioncactus"),
        "artifact-contract.md names the Pretendard stylesheet"
    );
    assert!(
        !text.contains("mermaid.min.js"),
        "artifact-contract.md names a mermaid script"
    );
}

#[test]
fn the_delivery_section_follows_the_report_rule() {
    let skill = read("SKILL.md");
    let delivery = delivery_section(&skill);
    let steps = [
        "**The md record.**",
        "**The html beside it.**",
        "**The Artifact publish, once.**",
        "**The refusal line.**",
    ];
    let mut last = 0;
    for step in steps {
        let at = delivery
            .find(step)
            .unwrap_or_else(|| panic!("the delivery section does not name {step}"));
        assert!(at >= last, "{step} is out of order");
        last = at;
    }
    assert!(delivery.contains("~/.moai/reports/"), "{delivery}");
    // Every `references/...` file SKILL.md points to ships: in this skill or in the moai skill.
    let root = Path::new(env!("CARGO_MANIFEST_DIR")).join("skills");
    // A path with a `<placeholder>` is a pattern, not a file.
    let named: Vec<&str> = skill
        .match_indices("`references/")
        .filter_map(|(i, _)| skill[i + 1..].split('`').next())
        .filter(|rel| !rel.contains('<'))
        .collect();
    assert!(
        named.contains(&"references/artifact-contract.md"),
        "{named:?}"
    );
    for rel in named {
        let ships = |skill_dir: &str| {
            let path = root.join(skill_dir).join(rel);
            if rel.ends_with('/') {
                path.is_dir()
            } else {
                path.is_file()
            }
        };
        assert!(
            ships("moai-html-report") || ships("moai"),
            "SKILL.md points to {rel}, which no shipped skill holds"
        );
    }
    assert!(
        !skill.contains("xdg-open"),
        "SKILL.md still opens a browser"
    );
}
