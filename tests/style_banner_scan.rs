//! The output style banners carry the moai mark: `<emoji> MoAI 🗿 <name> ─…` (user instruction
//! 2026-11-21). Nothing shipped keeps the old `MoAI-Easy ★` or `MoAI ★` wordings, no bare `★`
//! banner marker remains in the shipped styles, and the style references pin the new form.
//! Walks the shipped tree like the other repo scans; listed in tools/repo_scans.json.

mod style_banner {
    use std::{
        fs,
        path::{Path, PathBuf},
    };

    const ROOT: &str = env!("CARGO_MANIFEST_DIR");
    const OLD_BANNERS: [&str; 2] = ["MoAI-Easy ★", "MoAI ★"];

    fn files(dir: &Path, out: &mut Vec<PathBuf>) {
        for e in fs::read_dir(dir).unwrap() {
            let p = e.unwrap().path();
            if p.is_dir() {
                files(&p, out);
            } else {
                out.push(p);
            }
        }
    }

    #[test]
    fn shipped_paths_keep_no_old_banner_wording() {
        let root = Path::new(ROOT);
        // `templates/` holds the shipped style templates; `skills/moai/` is the shipped skill
        // package whose references state the banner.
        let mut shipped = Vec::new();
        for dir in ["templates", "skills/moai"] {
            files(&root.join(dir), &mut shipped);
        }
        assert!(
            shipped.len() > 30,
            "scan found too few files: {}",
            shipped.len()
        );
        let mut offenders = Vec::new();
        for path in shipped {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let hits: Vec<&str> = OLD_BANNERS
                .into_iter()
                .filter(|old| text.contains(old))
                .collect();
            if !hits.is_empty() {
                offenders.push(format!("{}: {}", path.display(), hits.join(", ")));
            }
        }
        assert!(
            offenders.is_empty(),
            "old banner wording or bare star marker in shipped paths:\n{}",
            offenders.join("\n")
        );
    }

    /// SPEC-20261124-card-flow-icons REQ-10: no shipped text (the templates, the style texts
    /// among them, and the skill package with its agents and references) holds an icon of the
    /// earlier card flow; the table `moai_cli::statusline::icons` names them.
    #[test]
    fn no_shipped_text_holds_an_icon_of_the_earlier_card_flow() {
        let root = Path::new(ROOT);
        let mut shipped = Vec::new();
        for dir in ["templates", "skills/moai"] {
            files(&root.join(dir), &mut shipped);
        }
        assert!(shipped.len() > 30, "scan found too few files");
        let mut offenders = Vec::new();
        for path in shipped {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            let hits: Vec<&str> = moai_cli::statusline::icons::RETIRED_FLOW
                .into_iter()
                .filter(|icon| text.contains(icon))
                .collect();
            if !hits.is_empty() {
                offenders.push(format!("{}: {}", path.display(), hits.join(" ")));
            }
        }
        assert!(
            offenders.is_empty(),
            "icons of the earlier card flow in shipped text:\n{}",
            offenders.join("\n")
        );
    }

    #[test]
    fn the_style_references_pin_the_new_banner_form() {
        for rel in [
            "skills/moai/references/home/output-style.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            assert!(text.contains("MoAI 🗿"), "{rel} lost the new banner mark");
            // every banner mark in a shipped style is the moai emoji: a bare star left in a
            // style file is a banner the change missed
            assert!(!text.contains('★'), "{rel} still carries a star marker");
        }
    }

    /// The card-flow icon set of the progress board, the status line and the lane reports is the
    /// table `moai_cli::statusline::icons` (SPEC-20261124-card-flow-icons): the three templates
    /// spell the flow sentence and the beside sentence the table builds, and carry none of the
    /// icons that left the set.
    #[test]
    fn the_three_style_templates_use_the_one_icon_set() {
        let flow = moai_cli::statusline::icons::flow_sentence();
        let beside = moai_cli::statusline::icons::beside_sentence();
        for rel in [
            "templates/claude/output-styles/moai.md",
            "templates/claude/output-styles/moai-easy.md",
            "templates/claude/output-styles/moai-learn.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            for old in moai_cli::statusline::icons::RETIRED {
                assert!(
                    !text.contains(old),
                    "{rel} still carries the old icon {old}"
                );
            }
            assert!(text.contains(&flow), "{rel} does not spell `{flow}`");
            assert!(text.contains(&beside), "{rel} does not spell `{beside}`");
        }
    }

    /// The name the style carried before SPEC-20261127-output-style-moai-style.
    const OLD_NAME: &str = "MoAI-Easy";

    /// Every file that may still hold the earlier name, each with the reason it keeps it.
    const OLD_NAME_ALLOWED: [(&str, &str); 27] = [
        (
            "src/home_style.rs",
            "the move's source: the earlier file name and value `moai update` moves away from",
        ),
        (
            "src/claude.rs",
            "the Go generation wrote this name; read-only detection of what it left",
        ),
        (
            "src/claude_settings_merge.rs",
            "the Go generation wrote this name; read-only detection of what it left",
        ),
        (
            "src/go_state.rs",
            "the Go generation wrote this name; read-only detection of what it left",
        ),
        (
            "tests/home_style_cli.rs",
            "tests of the move: the earlier name is its source",
        ),
        (
            "tests/home_style_ownership_cli.rs",
            "tests of the move and of the earlier renders",
        ),
        (
            "tests/update_user_layer_cli.rs",
            "the `moai update` test of the move",
        ),
        ("tests/go_state_cli.rs", "Go generation fixture"),
        ("tests/go_state_preview_cli.rs", "Go generation fixture"),
        ("tests/go_project_real_run_cli.rs", "Go generation fixture"),
        ("tests/project_update_plan_cli.rs", "Go generation fixture"),
        (
            "tests/style_banner_scan.rs",
            "this scan, and the forbidden star banner",
        ),
        (
            "tests/fixtures/home-style/render-1.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-2.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-3.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-4.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-5.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-6.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-7.md",
            "an earlier render, pinned by its sha256",
        ),
        (
            "tests/fixtures/home-style/render-8.md",
            "the render before this change, pinned by its sha256",
        ),
        (
            "tests/fixtures/atlas/source_snapshot.json",
            "a snapshot of the Go generation detection source",
        ),
        (
            "skills/moai-html-report/SKILL.md",
            "the tier map keeps the earlier name for sessions started before the rename",
        ),
        (
            "templates/bundle-packs/.claude/skills/moai-domain-html-report/SKILL.md",
            "the same tier map in the bundle-pack copy",
        ),
        (
            "templates/claude/output-styles/moai-easy.md",
            "upstream adapted template, not renamed (D-9)",
        ),
        (
            "templates/claude/output-styles/NOTICE.md",
            "names the upstream adapted templates (D-9)",
        ),
        (
            "tools/repo_scans.json",
            "the forbidden banner wording `MoAI-Easy ★`",
        ),
        (
            "tests/fixtures/statusline/stdin/a05-pr-approved-effort.json",
            "a recorded Claude Code status line input and its goldens",
        ),
    ];

    /// AC-12 (REQ-1, REQ-21): the earlier name stays only in the files of the list above; a
    /// listed file that no longer holds it is a stale entry. The recorded status line goldens and
    /// update screens are matched by prefix.
    #[test]
    fn ac12_the_old_style_name_stays_only_where_it_is_the_move_source_or_history() {
        let root = Path::new(ROOT);
        let mut all = Vec::new();
        for dir in ["src", "tests", "skills", "templates", "tools"] {
            files(&root.join(dir), &mut all);
        }
        // The Go generation helper is matched by prefix: spelling its full path in this file would
        // make the Windows gate scan read this file as an includer of the helper.
        let by_prefix = [
            "tests/fixtures/statusline/golden/a05-pr-approved-effort.",
            "tests/fixtures/update_tux/preview-",
            "tests/common/go_project_real_run.",
        ];
        let mut seen = std::collections::BTreeSet::new();
        for path in all {
            let Ok(text) = fs::read_to_string(&path) else {
                continue;
            };
            if text.contains(OLD_NAME) {
                let rel = path
                    .strip_prefix(root)
                    .unwrap()
                    .to_string_lossy()
                    .replace('\\', "/");
                seen.insert(rel);
            }
        }
        let listed: std::collections::BTreeSet<String> = OLD_NAME_ALLOWED
            .iter()
            .map(|(p, why)| {
                assert!(!why.is_empty(), "{p} has no reason");
                (*p).to_string()
            })
            .collect();
        let unlisted: Vec<&String> = seen
            .iter()
            .filter(|p| !listed.contains(*p) && !by_prefix.iter().any(|b| p.starts_with(b)))
            .collect();
        assert!(
            unlisted.is_empty(),
            "the earlier style name `{OLD_NAME}` is in files not on the list:\n{}",
            unlisted
                .iter()
                .map(|p| p.as_str())
                .collect::<Vec<_>>()
                .join("\n")
        );
        let stale: Vec<&String> = listed.iter().filter(|p| !seen.contains(*p)).collect();
        assert!(stale.is_empty(), "stale entries on the list: {stale:?}");
        for prefix in by_prefix {
            assert!(
                seen.iter().any(|p| p.starts_with(prefix)),
                "no file with the prefix {prefix}"
            );
        }
    }

    /// AC-13 (REQ-21): the tier map of the html-report skill and of its bundle-pack copy resolves
    /// `MoAI-Style` to `basic` and keeps `MoAI-Easy` resolving to `basic` as the earlier name.
    #[test]
    fn ac13_the_html_report_tier_map_knows_the_new_and_the_earlier_style_name() {
        for rel in [
            "skills/moai-html-report/SKILL.md",
            "templates/bundle-packs/.claude/skills/moai-domain-html-report/SKILL.md",
        ] {
            let text = fs::read_to_string(Path::new(ROOT).join(rel)).unwrap();
            for name in ["MoAI-Style", "MoAI-Easy"] {
                let row = format!("| `{name}` | `basic` |");
                assert_eq!(text.matches(&row).count(), 1, "{rel}: one `{row}` row");
            }
            assert!(
                text.contains("With `MoAI-Style` active, this resolves to the `basic` tier"),
                "{rel}: the worked example names the new style"
            );
        }
    }
}
