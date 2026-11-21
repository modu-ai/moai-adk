//! Wait for the next clock second instead of sleeping a fixed 1100 ms.
//!
//! Receipts, backup stamps and index generations carry a one-second timestamp, and the product
//! refuses two installs that finish in the same second. A test that runs a second command
//! "later" therefore needs the clock to have moved on by one second, not 1.1 seconds of sleep:
//! the first command ran in a second no later than the one the clock shows now, so once the
//! clock shows a different second, the next command runs in a strictly later one. The wait
//! polls the clock and ends within about a second (half a second on average).
//!
//! `tests/common/mod.rs` is left alone on purpose: every test crate includes it, and a change
//! there makes `moai check affected` select nearly every test.
#![allow(dead_code)]

use std::time::{Duration, Instant, SystemTime, UNIX_EPOCH};

/// The longest the clock may stay on one second before the wait gives up.
pub const NEXT_SECOND_LIMIT: Duration = Duration::from_secs(3);

const POLL: Duration = Duration::from_millis(5);

/// The Unix second the system clock shows now.
pub fn unix_second() -> u64 {
    SystemTime::now()
        .duration_since(UNIX_EPOCH)
        .expect("the system clock is after 1970")
        .as_secs()
}

/// Block until the system clock shows a later Unix second than it did at the call.
///
/// Fails the test, naming this wait, when the second does not change within
/// [`NEXT_SECOND_LIMIT`] (a clock that was set back, or a stopped clock).
pub fn wait_for_next_second() {
    wait_for_next_second_within(NEXT_SECOND_LIMIT, unix_second);
}

/// [`wait_for_next_second`] with its clock and limit given, so the failure path can be run.
pub fn wait_for_next_second_within(limit: Duration, clock: impl Fn() -> u64) {
    let start = clock();
    let began = Instant::now();
    while clock() <= start {
        assert!(
            began.elapsed() < limit,
            "wait_for_next_second: the clock stayed at Unix second {start} for {limit:?}"
        );
        std::thread::sleep(POLL);
    }
}
