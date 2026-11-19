//! HeadVer property tests (C-100, SPEC-20261119-property-tests): the compare on formatted
//! versions is a total order consistent with the components and `is_newer`. The shared seeded
//! generator's own tests live here, once, because `#[path]` would compile them into every
//! suite that includes `tests/common/prop.rs`.

#[path = "common/prop.rs"]
mod prop;

use moai_cli::headver;
use prop::{run_property, Rng, MIN_CASES};
use std::cmp::Ordering;

/// Generated components: a head of the valid range, a four-digit yearweek that stays a
/// plausible ISO week number, and a build counter.
fn components(rng: &mut Rng) -> (u32, u32, u32) {
    let head = rng.below(10_000) as u32;
    let yearweek = rng.below(100) as u32 * 100 + 1 + rng.below(53) as u32;
    let build = rng.below(1_000_000) as u32;
    (head, yearweek, build)
}

/// REQ-3, REQ-4: for two formatted versions `compare` answers exactly the component order,
/// symmetrically, self-equal, with `is_newer` agreeing.
#[test]
fn compare_orders_formatted_versions_as_their_components_do() {
    run_property("headver order", MIN_CASES, |rng| {
        let (h1, y1, b1) = components(rng);
        let (h2, y2, b2) = components(rng);
        let v1 = headver::version(h1, y1, b1);
        let v2 = headver::version(h2, y2, b2);
        let got = headver::compare(&v1, &v2)
            .unwrap_or_else(|| panic!("compare answered None for {v1} vs {v2}"));
        let want = (h1, y1, b1).cmp(&(h2, y2, b2));
        assert_eq!(got, want, "{v1} vs {v2}");
        assert_eq!(
            headver::compare(&v2, &v1)
                .unwrap_or_else(|| panic!("compare answered None for {v2} vs {v1}")),
            got.reverse(),
            "symmetry: {v1} vs {v2}"
        );
        assert_eq!(
            headver::compare(&v1, &v1)
                .unwrap_or_else(|| panic!("compare answered None for {v1} vs itself")),
            Ordering::Equal
        );
        assert_eq!(headver::is_newer(&v1, &v2), got == Ordering::Greater);
        assert_eq!(headver::is_newer(&v2, &v1), got == Ordering::Less);
    });
}

/// REQ-3, REQ-4: a batch sorted by `compare` holds no inversion and never contradicts the
/// component tuples (transitivity made visible).
#[test]
fn sorted_batches_hold_no_inversion() {
    run_property("headver batch", MIN_CASES, |rng| {
        let mut batch: Vec<((u32, u32, u32), String)> = (0..8)
            .map(|_| {
                let c = components(rng);
                (c, headver::version(c.0, c.1, c.2))
            })
            .collect();
        batch.sort_by(|a, b| {
            headver::compare(&a.1, &b.1)
                .unwrap_or_else(|| panic!("compare answered None for {} vs {}", a.1, b.1))
        });
        for pair in batch.windows(2) {
            let ((ta, va), (tb, vb)) = (&pair[0], &pair[1]);
            let ord = headver::compare(va, vb)
                .unwrap_or_else(|| panic!("compare answered None for {va} vs {vb}"));
            assert_ne!(ord, Ordering::Greater, "{va} before {vb}");
            assert!(ta <= tb, "sorted {va} sits before {vb} but {ta:?} > {tb:?}");
        }
    });
}

/// REQ-1, REQ-2 (the harness's own tests, D-4): the generator is deterministic for a seed,
/// names stay in the path alphabet, a shuffle keeps the multiset, and a false property fails
/// naming its seed and case index (a canary must not pass silently).
#[test]
fn the_generator_and_the_canary() {
    let mut a = Rng::new(42);
    let mut b = Rng::new(42);
    for _ in 0..100 {
        assert_eq!(a.next_u64(), b.next_u64());
    }
    let mut rng = Rng::new(7);
    let name = rng.name(12);
    assert!(!name.is_empty() && name.len() <= 12);
    assert!(name
        .bytes()
        .all(|b| b.is_ascii_alphanumeric() || matches!(b, b'-' | b'_' | b'.')));
    let items: Vec<u32> = vec![1, 2, 3, 4, 5];
    let mut shuffled = rng.shuffled(&items);
    shuffled.sort_unstable();
    assert_eq!(shuffled, items);

    let caught = std::panic::catch_unwind(|| run_property("canary", 4, |_| panic!("boom")));
    let err = caught.expect_err("a false property must fail");
    let text = err
        .downcast_ref::<String>()
        .cloned()
        .or_else(|| err.downcast_ref::<&str>().map(|s| s.to_string()))
        .unwrap();
    assert!(text.contains("property canary failed at case 0"), "{text}");
    assert!(text.contains("PROP_SEED="), "{text}");
}
