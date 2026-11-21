//! Tests that need a later clock second wait for the clock, not for 1100 ms of sleep.
//!
//! The wait helper (`tests/clock_wait/mod.rs`) is run here with a real clock and with fake
//! ones, and the test files that used to sleep a fixed 1100 ms are read to show they no longer
//! do and call the helper instead.

mod clock_wait;

use clock_wait::{unix_second, wait_for_next_second, wait_for_next_second_within};
use std::{
    cell::Cell,
    panic::{catch_unwind, AssertUnwindSafe},
    time::{Duration, Instant},
};

/// The test files whose fixed 1100 ms sleep became a wait on the clock.
const WAITING_FILES: [(&str, &str); 5] = [
    (
        "go_state_classes_cli.rs",
        include_str!("go_state_classes_cli.rs"),
    ),
    ("index_cli.rs", include_str!("index_cli.rs")),
    (
        "update_command_cli.rs",
        include_str!("update_command_cli.rs"),
    ),
    (
        "update_moved_receipts.rs",
        include_str!("update_moved_receipts.rs"),
    ),
    (
        "update_version_folders.rs",
        include_str!("update_version_folders.rs"),
    ),
];

fn panic_message(result: std::thread::Result<()>) -> String {
    let payload = result.expect_err("the wait should have failed");
    payload
        .downcast_ref::<String>()
        .cloned()
        .or_else(|| payload.downcast_ref::<&str>().map(|s| s.to_string()))
        .unwrap_or_default()
}

#[test]
fn the_clock_wait_ends_when_the_second_changes_and_fails_when_it_does_not() {
    // A real clock: the wait ends in a later second and within about one second.
    let before = unix_second();
    let began = Instant::now();
    wait_for_next_second();
    let took = began.elapsed();
    assert!(unix_second() > before, "the clock did not move on");
    assert!(
        took < Duration::from_millis(1100),
        "the wait took {took:?}, as long as the fixed sleep it replaces"
    );

    // A fake clock that moves on at its fourth reading ends the wait.
    let reads = Cell::new(0u64);
    wait_for_next_second_within(Duration::from_secs(1), || {
        reads.set(reads.get() + 1);
        100 + reads.get() / 4
    });

    // A stopped clock fails the test and the message names the wait and the second.
    let stopped = catch_unwind(AssertUnwindSafe(|| {
        wait_for_next_second_within(Duration::from_millis(60), || 4_000_000_000);
    }));
    let message = panic_message(stopped);
    assert!(message.contains("wait_for_next_second"), "{message}");
    assert!(message.contains("4000000000"), "{message}");

    // A clock set back during the wait is no later second: it fails the same way.
    let ticks = Cell::new(10u64);
    let set_back = catch_unwind(AssertUnwindSafe(|| {
        wait_for_next_second_within(Duration::from_millis(60), || {
            ticks.set(ticks.get().saturating_sub(1));
            ticks.get()
        });
    }));
    assert!(panic_message(set_back).contains("wait_for_next_second"));
}

#[test]
fn the_listed_files_hold_no_fixed_1100_ms_sleep() {
    for (name, text) in WAITING_FILES {
        assert!(
            !text.contains("from_millis(1100)"),
            "{name} sleeps a fixed 1100 ms; wait for the next clock second instead"
        );
        assert!(
            text.contains("clock_wait::wait_for_next_second()"),
            "{name} does not wait on the clock"
        );
    }
}
