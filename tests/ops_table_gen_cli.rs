//! SPEC-20261119-sorted-registration (card C-75, plan slice): pins the measured plan. The
//! research document must hold the ten-plus historical merge pairs with their before/after
//! conflict counts (no after worse than a before) and the three conclusions (C-1 docs-table
//! recommendation, C-2 sorting-not-supported, C-3 already-solved). Reads only the SPEC folder;
//! writes nothing.

use std::path::Path;

fn research() -> String {
    let path = Path::new(env!("CARGO_MANIFEST_DIR"))
        .join(".moai/specs/SPEC-20261119-sorted-registration/research.md");
    std::fs::read_to_string(path).expect("read research.md")
}

/// AC-1: at least ten measured pairs, every after count no worse than the before count.
#[test]
fn the_plan_holds_ten_pairs_whose_after_never_worsens_the_before() {
    let text = research();
    let mut pairs = 0;
    for line in text
        .lines()
        .filter(|l| l.starts_with("| ") && !l.contains("---"))
    {
        let cells: Vec<&str> = line.split('|').skip(1).map(str::trim).collect();
        // A pair row: merge(12 hex) | file | before | after(sorted) | after(rows removed)|?
        if cells.len() >= 4 && cells[0].len() == 12 {
            let before: usize = cells[2].parse().expect("before is a number");
            for after in cells[3..].iter().filter_map(|c| c.parse::<usize>().ok()) {
                assert!(
                    after <= before,
                    "pair {} got worse: {before} -> {after}",
                    cells[0]
                );
            }
            pairs += 1;
        }
    }
    assert!(pairs >= 10, "the plan measured only {pairs} pairs");
}

/// AC-1: the conclusions name the three findings the measurements support.
#[test]
fn the_conclusions_name_the_recommendation_the_sorting_result_and_the_solved_part() {
    let text = research();
    assert!(
        text.contains("**C-1 (recommended): the operation table must not be committed.**"),
        "C-1 recommendation missing"
    );
    assert!(
        text.contains("**C-2: sorting `src/lib.rs` and `src/main.rs` lists is not supported by measurement.**"),
        "C-2 sorting result missing"
    );
    assert!(
        text.contains("**C-3: what already works stays.**"),
        "C-3 already-solved missing"
    );
    // The measured gain C-1 claims is in the numbers: 7 of 8 table pairs.
    assert!(text.contains("7 of 8"), "the C-1 gain is not stated");
}

/// AC-1: every measured pair names its merge commit and the file it conflicts in.
#[test]
fn every_pair_names_a_merge_commit_and_a_shared_file() {
    let text = research();
    let shared = ["src/main.rs", "src/lib.rs", "docs/rust-product-path.md"];
    let mut counted = 0;
    for line in text
        .lines()
        .filter(|l| l.starts_with("| ") && !l.contains("---"))
    {
        let cells: Vec<&str> = line.split('|').skip(1).map(str::trim).collect();
        if cells.len() >= 4 && cells[0].len() == 12 {
            assert!(
                cells[0].chars().all(|c| c.is_ascii_hexdigit()),
                "{} is not a commit sha",
                cells[0]
            );
            assert!(
                shared.contains(&cells[1]),
                "{} is not one of the three shared files",
                cells[1]
            );
            counted += 1;
        }
    }
    assert!(counted >= 10, "only {counted} rows name a commit and file");
}
