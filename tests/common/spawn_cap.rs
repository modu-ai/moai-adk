//! A bounded set of permits for the tests that spawn the moai binary (REQ-3 of
//! SPEC-20261120-test-parallelism-cap): one test binary never runs more spawning tests at
//! once than `MOAI_SPAWN_TEST_CAP` (default 4), so a whole run keeps its process count
//! bounded. A test takes one permit around its spawn and releases it on drop; a test that
//! never spawns never calls this. In-process on purpose (SPEC decision D3): the card bounds
//! one test binary, and a cross-binary lock would add contention for no named requirement.
//!
//! A test file includes this with `#[path = "common/spawn_cap.rs"] mod spawn_cap;` — the
//! shared `mod.rs` is not touched, so a file that does not need the cap stays untouched.

use std::sync::{Condvar, Mutex, OnceLock};

/// The environment variable that names the cap, read once per test binary.
pub const ENV: &str = "MOAI_SPAWN_TEST_CAP";

const DEFAULT_CAP: usize = 4;

struct Permits {
    left: Mutex<usize>,
    freed: Condvar,
}

fn cap() -> usize {
    static CAP: OnceLock<usize> = OnceLock::new();
    *CAP.get_or_init(|| {
        std::env::var(ENV)
            .ok()
            .and_then(|v| v.trim().parse::<usize>().ok())
            .filter(|n| *n > 0)
            .unwrap_or(DEFAULT_CAP)
    })
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
pub struct SpawnPermit;

impl Drop for SpawnPermit {
    fn drop(&mut self) {
        let p = permits();
        let mut left = p.left.lock().unwrap_or_else(|e| e.into_inner());
        *left += 1;
        drop(left);
        p.freed.notify_one();
    }
}

/// Takes one spawn permit, waiting while every permit of the cap is held.
pub fn acquire() -> SpawnPermit {
    let p = permits();
    let mut left = p.left.lock().unwrap_or_else(|e| e.into_inner());
    while *left == 0 {
        left = p.freed.wait(left).unwrap_or_else(|e| e.into_inner());
    }
    *left -= 1;
    SpawnPermit
}
