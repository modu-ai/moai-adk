//! A bounded set of permits for the tests that spawn the moai binary (REQ-3 of
//! SPEC-20261120-test-parallelism-cap): one test binary never runs more spawning tests at
//! once than `MOAI_SPAWN_TEST_CAP` (default 4), so a whole run keeps its process count
//! bounded. A test takes one permit around its spawn and releases it on drop; a test that
//! never spawns never calls this. In-process on purpose (SPEC decision D3): the card bounds
//! one test binary, and a cross-binary lock would add contention for no named requirement.
//!
//! A test file includes this with `#[path = "common/spawn_cap.rs"] mod spawn_cap;` — the
//! shared `mod.rs` is not touched, so a file that does not need the cap stays untouched.

//!
//! Under `cargo nextest`, where every test is a process of its own, this in-process set bounds
//! nothing across tests: `.config/nextest.toml` carries the same bound as a test group
//! (`spawning-moai`, `max-threads = 4`) for the files that use this helper.

use std::cell::Cell;
use std::marker::PhantomData;
use std::sync::{Condvar, Mutex, OnceLock};

/// The environment variable that names the cap, read once per test binary.
pub const ENV: &str = "MOAI_SPAWN_TEST_CAP";

pub const DEFAULT_CAP: usize = 4;

struct Permits {
    left: Mutex<usize>,
    freed: Condvar,
}

/// The cap a variable's text names: a whole number of at least 1 (spaces around it are
/// ignored); no variable, an empty one, 0, a negative or fractional number, text, or a number
/// too large for `usize` is the default.
pub fn parse_cap(text: Option<&str>) -> usize {
    text.and_then(|v| v.trim().parse::<usize>().ok())
        .filter(|n| *n > 0)
        .unwrap_or(DEFAULT_CAP)
}

fn cap() -> usize {
    static CAP: OnceLock<usize> = OnceLock::new();
    *CAP.get_or_init(|| parse_cap(std::env::var(ENV).ok().as_deref()))
}

thread_local! {
    /// Whether this thread holds a permit: a second `acquire` from it would wait for itself
    /// (with a cap of 1, forever), so it panics and says so instead.
    static HELD: Cell<bool> = const { Cell::new(false) };
}

fn permits() -> &'static Permits {
    static PERMITS: OnceLock<Permits> = OnceLock::new();
    PERMITS.get_or_init(|| Permits {
        left: Mutex::new(cap()),
        freed: Condvar::new(),
    })
}

/// One permit to spawn the binary. Hold it across the spawn and the wait; the semaphore
/// bounds how many of these the binary holds at once.
#[must_use = "a permit is acquired for the spawn that follows; dropping it right away releases it unused"]
pub struct SpawnPermit(
    // Not `Send`: the permit is released on the thread that took it, where `HELD` was set.
    PhantomData<*const ()>,
);

impl Drop for SpawnPermit {
    fn drop(&mut self) {
        HELD.with(|h| h.set(false));
        let p = permits();
        let mut left = p.left.lock().unwrap_or_else(|e| e.into_inner());
        *left += 1;
        drop(left);
        p.freed.notify_one();
    }
}

/// Takes one spawn permit, waiting while every permit of the cap is held. Panics when the
/// calling thread already holds one: take it around one spawn and let it go before the next.
pub fn acquire() -> SpawnPermit {
    assert!(
        !HELD.with(Cell::get),
        "spawn_cap: nested acquire on one thread — the first permit must be dropped before a \
         helper that takes its own (a cap of 1 would wait forever)"
    );
    let p = permits();
    let mut left = p.left.lock().unwrap_or_else(|e| e.into_inner());
    while *left == 0 {
        left = p.freed.wait(left).unwrap_or_else(|e| e.into_inner());
    }
    *left -= 1;
    HELD.with(|h| h.set(true));
    SpawnPermit(PhantomData)
}
