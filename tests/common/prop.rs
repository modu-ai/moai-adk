//! The seeded in-house generator the property suites share (C-100, SPEC-20261119-property-tests).
//!
//! No new crate: one small SplitMix64 generator, every property names its seed, and a failure
//! prints the seed and the case index so the exact case can be replayed with
//! `PROP_SEED=<seed> PROP_ONLY=<i> cargo test --test <suite> <property>`.
//!
//! Each integration test includes this file directly, because `tests/common/mod.rs` is outside
//! the card's write scope:
//!
//! ```text
//! #[path = "common/prop.rs"]
//! mod prop;
//! ```
#![allow(dead_code)]

use std::panic::{catch_unwind, AssertUnwindSafe};
use std::time::{Duration, Instant};

/// One deterministic generator; `SplitMix64` seeded from the property's seed plus the case
/// index, so every case re-derives from one number.
pub struct Rng {
    state: u64,
}

impl Rng {
    pub fn new(seed: u64) -> Self {
        Self { state: seed }
    }

    pub fn next_u64(&mut self) -> u64 {
        self.state = self.state.wrapping_add(0x9e37_79b9_7f4a_7c15);
        let mut z = self.state;
        z = (z ^ (z >> 30)).wrapping_mul(0xbf58_476d_1ce4_e5b9);
        z = (z ^ (z >> 27)).wrapping_mul(0x94d0_49bb_1331_11eb);
        z ^ (z >> 31)
    }

    /// A value below `n` (the caller keeps `n` above zero).
    pub fn below(&mut self, n: u64) -> u64 {
        self.next_u64() % n
    }

    pub fn bool(&mut self) -> bool {
        self.next_u64() & 1 == 1
    }

    /// One of `items` (the caller keeps the list non-empty).
    pub fn pick<'a, T>(&mut self, items: &'a [T]) -> &'a T {
        &items[self.below(items.len() as u64) as usize]
    }

    /// A short name of the characters a project path may hold: ASCII letters, digits and
    /// `-`, `_`, `.` (the profile's `recordable` alphabet), 1..=`max` bytes, never a name of
    /// dots only (`.`, `..`, `...` must stay path steps, not names).
    pub fn name(&mut self, max_len: usize) -> String {
        const ALPHABET: &[u8] = b"abcdefghijklmnopqrstuvwxyz0123456789-_.";
        let len = 1 + self.below(max_len as u64) as usize;
        let mut name: String = (0..len)
            .map(|_| ALPHABET[self.below(ALPHABET.len() as u64) as usize] as char)
            .collect();
        if !name.bytes().any(|b| b.is_ascii_alphanumeric()) {
            name.insert(0, 'x');
        }
        name
    }

    /// A shuffled copy of `items`.
    pub fn shuffled<T: Clone>(&mut self, items: &[T]) -> Vec<T> {
        let mut out = items.to_vec();
        for i in (1..out.len()).rev() {
            out.swap(i, self.below((i + 1) as u64) as usize);
        }
        out
    }
}

/// The default seed of every property: fixed, so a clean run is reproducible; `PROP_SEED`
/// overrides it and `PROP_ONLY` narrows a replay to one case index.
fn seed_from_env() -> (u64, Option<u64>) {
    let seed = std::env::var("PROP_SEED")
        .ok()
        .and_then(|s| s.parse::<u64>().ok())
        .unwrap_or(0x0C10_2026_1119);
    let only = std::env::var("PROP_ONLY")
        .ok()
        .and_then(|s| s.parse::<u64>().ok());
    (seed, only)
}

/// Runs `property` for `cases` generated inputs and asserts the run stays within its budget
/// (C-100 AC-3: at least 1000 cases, 30 seconds per property, the seed printed on failure).
/// A failing case's panic is retold with the property's name, the seed and the case index.
/// `FnMut` so a property can hold one shared fixture (a ledger, a tree) across its cases.
pub fn run_property(name: &str, cases: u64, mut property: impl FnMut(&mut Rng)) {
    let (seed, only) = seed_from_env();
    let start = Instant::now();
    for i in 0..cases {
        if let Some(one) = only {
            if one != i {
                continue;
            }
        }
        let mut rng = Rng::new(seed ^ i.wrapping_mul(0x8000_0000_0000_0001 | 1));
        let caught = catch_unwind(AssertUnwindSafe(|| property(&mut rng)));
        if let Err(panic) = caught {
            let reason = panic
                .downcast_ref::<&str>()
                .map(|s| (*s).to_string())
                .or_else(|| panic.downcast_ref::<String>().cloned())
                .unwrap_or_else(|| "panic".to_string());
            panic!(
                "property {name} failed at case {i} of {cases} (seed {seed}): {reason} — replay \
                 with PROP_SEED={seed} PROP_ONLY={i}",
            );
        }
    }
    let load = None;
    if let Some(message) = over_budget(name, cases, seed, start.elapsed(), load, 1) {
        panic!("{message}");
    }
}

/// The time budget of one property run.
pub fn budget(_load: Option<f64>, _cpus: usize) -> Duration {
    Duration::from_secs(30)
}

/// The failure message of a run that took `elapsed`, or `None` while it is within the budget.
pub fn over_budget(
    name: &str,
    cases: u64,
    seed: u64,
    elapsed: Duration,
    load: Option<f64>,
    cpus: usize,
) -> Option<String> {
    (elapsed >= budget(load, cpus))
        .then(|| format!("property {name} ran {cases} cases in over 30 s (seed {seed})"))
}

/// The C-100 case floor: every property runs at least this many cases.
pub const MIN_CASES: u64 = 1000;

/// SPEC-20261122-test-defects-session-reclaim AC-7 (REQ-7): 30 s at or below one runnable
/// thread per CPU or with no reading, in proportion above it, and capped.
#[test]
fn the_property_budget_grows_with_load_above_the_cpus() {
    let secs = |load: Option<f64>, cpus: usize| budget(load, cpus).as_secs_f64();
    assert_eq!(secs(None, 8), 30.0);
    assert_eq!(secs(Some(2.0), 8), 30.0);
    assert_eq!(secs(Some(8.0), 8), 30.0);
    assert_eq!(secs(Some(16.0), 8), 60.0);
    assert_eq!(secs(Some(67.0), 16), 30.0 * 67.0 / 16.0);
    assert_eq!(secs(Some(1000.0), 8), 240.0, "capped at eight times");
    assert_eq!(secs(Some(f64::NAN), 8), 30.0);
    assert_eq!(secs(Some(-1.0), 8), 30.0);
    assert_eq!(secs(Some(16.0), 0), 30.0, "no CPU count, no scaling");
}

/// AC-7 (REQ-8): a run within its budget passes; one over it names the property, the seed, the
/// elapsed time, the budget and the load the budget was scaled by.
#[test]
fn the_property_budget_message_names_the_seed_and_the_load() {
    let within = over_budget("p", 1000, 7, Duration::from_secs(59), Some(16.0), 8);
    assert_eq!(within, None);
    let over = over_budget("p", 1000, 7, Duration::from_secs(61), Some(16.0), 8).expect("over");
    for part in [
        "property p",
        "1000 cases",
        "seed 7",
        "61",
        "60",
        "load 16.0",
        "8 CPUs",
    ] {
        assert!(over.contains(part), "{part:?} in {over}");
    }
}
