//! REQ-3 of SPEC-20261120-test-parallelism-cap: more threads than the cap go through the
//! permit helper at the same time, and the observed concurrency never exceeds the cap. The
//! cap is read once per test binary, so this file — its own test binary — sets the
//! environment variable to the same value before every acquire (whichever test runs first
//! initializes it) and bounds nothing outside itself.

#[path = "common/spawn_cap.rs"]
mod spawn_cap;

use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::{Arc, Barrier};

const CAP: &str = "2";

#[test]
fn three_threads_take_permits_and_never_hold_three() {
    std::env::set_var(spawn_cap::ENV, CAP);
    let held = Arc::new(AtomicUsize::new(0));
    let peak = Arc::new(AtomicUsize::new(0));
    let barrier = Arc::new(Barrier::new(3));
    let mut threads = Vec::new();
    for _ in 0..3 {
        let held = held.clone();
        let peak = peak.clone();
        let barrier = barrier.clone();
        threads.push(std::thread::spawn(move || {
            barrier.wait();
            let permit = spawn_cap::acquire();
            let now = held.fetch_add(1, Ordering::SeqCst) + 1;
            peak.fetch_max(now, Ordering::SeqCst);
            std::thread::sleep(std::time::Duration::from_millis(50));
            held.fetch_sub(1, Ordering::SeqCst);
            drop(permit);
        }));
    }
    for t in threads {
        t.join().unwrap();
    }
    let seen = peak.load(Ordering::SeqCst);
    assert!(
        seen <= 2,
        "the semaphore let {seen} threads hold permits at once"
    );
}

#[test]
fn a_dropped_permit_is_freed_for_the_next_taker() {
    std::env::set_var(spawn_cap::ENV, CAP);
    let first = spawn_cap::acquire();
    let second = spawn_cap::acquire();
    // The cap is fully held now: dropping one lets a third taker through — without the drop
    // freeing its permit this acquire would hang.
    drop(first);
    let third = spawn_cap::acquire();
    drop(second);
    drop(third);
}
