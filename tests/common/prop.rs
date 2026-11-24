//! The seeded in-house generator the property suites share (SPEC-20261119-property-tests).
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
    let only = replay_index(std::env::var("PROP_ONLY").ok().as_deref())
        .unwrap_or_else(|message| panic!("{message}"));
    (seed, only)
}

/// The replay index from the raw `PROP_ONLY` value (REQ-17): none when it is unset, the number
/// when it is one, and a message naming the value when it is set but is not a number, so a typo
/// never turns a replay into a silent full run.
pub fn replay_index(raw: Option<&str>) -> Result<Option<u64>, String> {
    match raw {
        None => Ok(None),
        Some(s) => s.parse::<u64>().map(Some).map_err(|_| {
            format!(
                "PROP_ONLY={s:?} is not a number; give the index of one case, or leave it unset"
            )
        }),
    }
}

/// The failure message of a replay that names a case the run does not have (REQ-13), or `None`
/// when `only` is absent or inside the run.
pub fn replay_error(name: &str, cases: u64, seed: u64, only: Option<u64>) -> Option<String> {
    only.filter(|&i| i >= cases).map(|i| {
        format!(
            "property {name} was asked to replay case {i} (PROP_ONLY={i}) but runs {cases} cases \
             (seed {seed}), so no case would run"
        )
    })
}

/// Runs `property` for `cases` generated inputs and asserts the run stays within its budget
/// (SPEC-20261119-property-tests AC-3: at least 1000 cases, 30 seconds per property, the seed
/// printed on failure).
/// A failing case's panic is retold with the property's name, the seed and the case index.
/// `FnMut` so a property can hold one shared fixture (a ledger, a tree) across its cases.
pub fn run_property(name: &str, cases: u64, property: impl FnMut(&mut Rng)) {
    let (seed, only) = seed_from_env();
    run_cases(name, cases, seed, only, property);
}

/// [`run_property`] with the seed and the replay index given, so a test can drive it without
/// touching the process environment.
fn run_cases(
    name: &str,
    cases: u64,
    seed: u64,
    only: Option<u64>,
    mut property: impl FnMut(&mut Rng),
) {
    if let Some(message) = replay_error(name, cases, seed, only) {
        panic!("{message}");
    }
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
    let elapsed = start.elapsed();
    // Read when the run ends, the moment the budget is checked.
    let cpus = std::thread::available_parallelism().map_or(0, |n| n.get());
    if let Some(message) = over_budget(name, cases, seed, elapsed, load_average(), cpus) {
        panic!("{message}");
    }
}

/// The budget on an idle machine (SPEC-20261119-property-tests AC-3).
const IDLE_BUDGET_SECS: f64 = 30.0;

/// The most the budget grows: eight times the idle budget.
const MAX_BUDGET_FACTOR: f64 = 8.0;

/// The time budget of one property run (SPEC-20261122-test-defects-session-reclaim REQ-7):
/// 30 s while the one-minute load is at or below the number of logical CPUs, or when either
/// cannot be read, and 30 s times load / CPUs above it, at most eight times. On a busy
/// machine the same cases take longer without any invariant being at stake.
pub fn budget(load: Option<f64>, cpus: usize) -> Duration {
    let factor = match load {
        Some(l) if l.is_finite() && cpus > 0 => (l / cpus as f64).clamp(1.0, MAX_BUDGET_FACTOR),
        _ => 1.0,
    };
    Duration::from_secs_f64(IDLE_BUDGET_SECS * factor)
}

/// The failure message of a run that took `elapsed`, or `None` while it is within the budget
/// (REQ-8): the property, the cases, the seed, the elapsed time, the budget and the load the
/// budget was scaled by.
pub fn over_budget(
    name: &str,
    cases: u64,
    seed: u64,
    elapsed: Duration,
    load: Option<f64>,
    cpus: usize,
) -> Option<String> {
    let budget = budget(load, cpus);
    (elapsed >= budget).then(|| {
        let load = load.map_or("unknown".to_string(), |l| format!("{l:.1}"));
        format!(
            "property {name} ran {cases} cases in {:.1} s, over its budget of {:.0} s (seed \
             {seed}, load {load} on {cpus} CPUs)",
            elapsed.as_secs_f64(),
            budget.as_secs_f64()
        )
    })
}

/// The one-minute load average, `None` where it cannot be read.
fn load_average() -> Option<f64> {
    #[cfg(unix)]
    {
        let mut loads = [0f64; 3];
        // SAFETY: `getloadavg` writes at most the one sample asked for into the array.
        let got = unsafe { libc::getloadavg(loads.as_mut_ptr(), 1) };
        (got == 1).then_some(loads[0])
    }
    #[cfg(not(unix))]
    {
        None
    }
}

/// The case floor of SPEC-20261119-property-tests: every property runs at least this many cases.
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

/// The text of a verdict that carries one, `None` for `Within`.
#[cfg(test)]
fn verdict_text(verdict: &Verdict) -> Option<&str> {
    match verdict {
        Verdict::Within => None,
        Verdict::Warn(text) | Verdict::Fail(text) => Some(text),
    }
}

/// SPEC-20261124-time-budget-tests AC-1 (REQ-1, REQ-3): the run of 2026-11-24 (51.7 s wall at
/// load 25.4 on 16 CPUs) used 15 s of CPU, so it passes with one warning line; the same wall
/// time with 31 s of CPU fails.
#[test]
fn the_property_budget_warns_not_fails_when_only_the_wall_time_passed_the_budget() {
    let name = "random_interleavings_keep_the_two_invariants";
    let wall = Duration::from_secs_f64(51.7);
    let cpu = Some(Duration::from_secs(15));
    let warned = judge(name, 1000, 13263398375705, wall, cpu, Some(25.4), 16);
    let Verdict::Warn(line) = &warned else {
        panic!("expected a warning, got {warned:?}");
    };
    for part in [name, "seed 13263398375705", "51.7", "48", "25.4", "15.0"] {
        assert!(line.contains(part), "{part:?} in {line}");
    }
    assert!(!line.contains('\n'), "one line: {line}");
    let failed = judge(
        name,
        1000,
        13263398375705,
        wall,
        Some(Duration::from_secs(31)),
        Some(25.4),
        16,
    );
    assert!(matches!(failed, Verdict::Fail(_)), "got {failed:?}");
}

/// AC-2 (REQ-2): 30.0 s of CPU fails and names the property, the cases, the seed, the CPU time
/// and the wall time; 29.9 s does not.
#[test]
fn the_property_budget_fails_at_thirty_seconds_of_cpu_time() {
    let wall = Duration::from_secs(40);
    let at = judge(
        "p",
        1000,
        7,
        wall,
        Some(Duration::from_secs_f64(30.0)),
        Some(16.0),
        8,
    );
    let Verdict::Fail(text) = &at else {
        panic!("expected a failure at 30.0 s, got {at:?}");
    };
    for part in ["property p", "1000 cases", "seed 7", "CPU", "30.0", "40.0"] {
        assert!(text.contains(part), "{part:?} in {text}");
    }
    let below = judge(
        "p",
        1000,
        7,
        Duration::from_secs(5),
        Some(Duration::from_secs_f64(29.9)),
        Some(1.0),
        8,
    );
    assert!(matches!(below, Verdict::Within), "got {below:?}");
}

/// AC-3 (REQ-4): 240 s of wall time fails as a hang whatever the CPU time and names the property,
/// the seed, the wall time and the CPU time; 239.9 s within a 240 s budget does not.
#[test]
fn the_property_budget_fails_a_hang_at_two_hundred_forty_seconds() {
    let hang = judge(
        "p",
        1000,
        7,
        Duration::from_secs(240),
        Some(Duration::from_secs(1)),
        Some(1000.0),
        8,
    );
    let Verdict::Fail(text) = &hang else {
        panic!("expected a hang, got {hang:?}");
    };
    for part in ["property p", "seed 7", "240.0", "1.0", "hang"] {
        assert!(text.contains(part), "{part:?} in {text}");
    }
    let no_reading = judge("p", 1000, 7, Duration::from_secs(240), None, None, 8);
    let text = match &no_reading {
        Verdict::Fail(text) => text,
        other => panic!("expected a hang with no CPU reading, got {other:?}"),
    };
    assert!(text.contains("unknown"), "unknown CPU time in {text}");
    let near = judge(
        "p",
        1000,
        7,
        Duration::from_secs_f64(239.9),
        Some(Duration::from_secs(1)),
        Some(1000.0),
        8,
    );
    assert_eq!(verdict_text(&near), None, "got {near:?}");
}

/// AC-4 (REQ-5): with no CPU reading, the wall time is judged against the load-scaled budget as
/// before: past it fails, within it passes.
#[test]
fn the_property_budget_uses_the_wall_time_when_no_cpu_time_is_read() {
    let over = judge("p", 1000, 7, Duration::from_secs(61), None, Some(16.0), 8);
    let Verdict::Fail(text) = &over else {
        panic!("expected a failure, got {over:?}");
    };
    for part in [
        "property p",
        "1000 cases",
        "seed 7",
        "61",
        "60",
        "load 16.0",
        "8 CPUs",
    ] {
        assert!(text.contains(part), "{part:?} in {text}");
    }
    let within = judge("p", 1000, 7, Duration::from_secs(59), None, Some(16.0), 8);
    assert!(matches!(within, Verdict::Within), "got {within:?}");
}

/// AC-5 (REQ-6): a case that panics fails the property with the name, the seed and the case
/// index, and the message never speaks of a budget.
#[test]
fn a_failing_case_fails_with_the_seed_and_case_index_whatever_the_time() {
    let mut calls = 0u64;
    let outcome = catch_unwind(AssertUnwindSafe(|| {
        run_cases("p", 5, 7, None, |_| {
            calls += 1;
            if calls == 3 {
                panic!("the third case broke");
            }
        })
    }));
    let panic = outcome.expect_err("a panicking case must fail the property");
    let text = panic
        .downcast_ref::<String>()
        .cloned()
        .or_else(|| panic.downcast_ref::<&str>().map(|s| (*s).to_string()))
        .expect("a text panic");
    for part in ["property p", "seed 7", "case 2", "the third case broke"] {
        assert!(text.contains(part), "{part:?} in {text}");
    }
    assert!(!text.contains("budget"), "no budget talk in {text}");
    assert_eq!(calls, 3, "the run stops at the failing case");
}

/// AC-6 (REQ-1): the thread CPU reading rises while the thread computes and stays near zero
/// while it sleeps.
#[cfg(unix)]
#[test]
fn the_thread_cpu_time_rises_when_the_thread_computes_and_not_when_it_sleeps() {
    let read = || thread_cpu_time().expect("the thread CPU time is readable on unix");
    let before = read();
    // The cap only stops a reading that never rises; it is far above the 20 ms asked for.
    let cap = Instant::now();
    let mut sink = 0u64;
    let after = loop {
        for i in 0..100_000u64 {
            sink = std::hint::black_box(sink.wrapping_add(i));
        }
        let now = read();
        if now - before >= Duration::from_millis(20) {
            break now;
        }
        assert!(
            cap.elapsed() < Duration::from_secs(120),
            "the thread CPU time did not rise by 20 ms while computing (sink {sink})"
        );
    };
    assert!(after - before >= Duration::from_millis(20));

    let slept_from = read();
    std::thread::sleep(Duration::from_millis(300));
    let slept_to = read();
    let rise = slept_to - slept_from;
    assert!(
        rise < Duration::from_millis(100),
        "the CPU reading rose by {rise:?} while the thread slept"
    );
}

/// AC-11 (REQ-13, REQ-14): a replay index at or beyond the case count would run no case and
/// pass; it fails and names the property, the index, the case count and the seed. An index
/// inside the run, or none, is accepted.
#[test]
fn a_replay_index_beyond_the_cases_is_refused() {
    assert_eq!(replay_error("p", 1000, 7, None), None);
    assert_eq!(replay_error("p", 1000, 7, Some(0)), None);
    assert_eq!(replay_error("p", 1000, 7, Some(999)), None);
    for beyond in [1000, 1001, u64::MAX] {
        let msg = replay_error("p", 1000, 7, Some(beyond)).expect("refused");
        for part in [
            "property p",
            &format!("PROP_ONLY={beyond}"),
            "1000 cases",
            "seed 7",
        ] {
            assert!(msg.contains(part), "{part:?} in {msg}");
        }
    }
}

/// AC-11 (REQ-14): the run itself refuses the index before it runs a case, and a replay inside
/// the run executes exactly that one case.
#[test]
fn the_run_refuses_a_replay_index_beyond_the_cases() {
    let mut ran = 0u64;
    let beyond = catch_unwind(AssertUnwindSafe(|| {
        run_cases("p", 3, 7, Some(5000), |_| ran += 1)
    }));
    assert!(beyond.is_err(), "an index beyond the cases passed");
    assert_eq!(ran, 0, "a case ran before the refusal");

    let mut ran = 0u64;
    run_cases("p", 3, 7, Some(1), |_| ran += 1);
    assert_eq!(ran, 1, "a replay inside the run runs one case");
}

/// AC-13 (REQ-17): a `PROP_ONLY` that is set but is not a number is refused and names the value;
/// a number, or no value, is accepted.
#[test]
fn a_replay_index_that_is_not_a_number_is_refused() {
    assert_eq!(replay_index(None), Ok(None));
    assert_eq!(replay_index(Some("0")), Ok(Some(0)));
    assert_eq!(replay_index(Some("999")), Ok(Some(999)));
    for bad in ["", "abc", "-1", "1.5", "12x"] {
        let msg = replay_index(Some(bad)).expect_err("refused");
        for part in ["PROP_ONLY", &format!("{bad:?}"), "number"] {
            assert!(msg.contains(part), "{part:?} in {msg}");
        }
    }
}
