//! REQ-3 of SPEC-20261120-test-parallelism-cap: more threads than the cap go through the
//! permit helper at the same time, and the observed concurrency never exceeds the cap. The
//! cap is read once per test binary, so this file — its own test binary — sets the
//! environment variable to the same value before every acquire (whichever test runs first
//! initializes it) and bounds nothing outside itself. The tests that take permits run one at
//! a time (`SERIAL`): they would otherwise take each other's permits and the observations
//! would measure the neighbour.

#[path = "common/spawn_cap.rs"]
mod spawn_cap;

use std::sync::atomic::{AtomicUsize, Ordering};
use std::sync::mpsc;
use std::sync::{Arc, Barrier, Mutex, MutexGuard};
use std::time::Duration;

const CAP: &str = "2";

static SERIAL: Mutex<()> = Mutex::new(());

fn serial() -> MutexGuard<'static, ()> {
    let guard = SERIAL.lock().unwrap_or_else(|e| e.into_inner());
    std::env::set_var(spawn_cap::ENV, CAP);
    guard
}

#[test]
fn three_threads_take_permits_and_never_hold_three() {
    let _serial = serial();
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
            std::thread::sleep(Duration::from_millis(50));
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

/// With the cap fully held a third taker waits, and a permit dropped by its holder frees it —
/// without the drop freeing its permit the third acquire would never return.
#[test]
fn a_dropped_permit_is_freed_for_the_next_taker() {
    let _serial = serial();
    let (got_tx, got_rx) = mpsc::channel::<()>();
    let mut release = Vec::new();
    let mut holders = Vec::new();
    for _ in 0..2 {
        let (tx, rx) = mpsc::channel::<()>();
        release.push(tx);
        let got_tx = got_tx.clone();
        holders.push(std::thread::spawn(move || {
            let permit = spawn_cap::acquire();
            got_tx.send(()).unwrap();
            rx.recv().unwrap();
            drop(permit);
        }));
    }
    for _ in 0..2 {
        got_rx.recv_timeout(Duration::from_secs(10)).unwrap();
    }
    let (third_tx, third_rx) = mpsc::channel::<()>();
    let third = std::thread::spawn(move || {
        let permit = spawn_cap::acquire();
        third_tx.send(()).unwrap();
        drop(permit);
    });
    assert!(
        third_rx.recv_timeout(Duration::from_millis(200)).is_err(),
        "a third permit was given while the cap was fully held"
    );
    release[0].send(()).unwrap();
    third_rx
        .recv_timeout(Duration::from_secs(10))
        .expect("a dropped permit did not free the waiting taker");
    release[1].send(()).unwrap();
    third.join().unwrap();
    for h in holders {
        h.join().unwrap();
    }
}

/// A permit dropped on a thread leaves that thread free to take another one.
#[test]
fn the_same_thread_can_take_a_permit_again_after_dropping_one() {
    let _serial = serial();
    drop(spawn_cap::acquire());
    drop(spawn_cap::acquire());
}

/// A second permit from a thread that holds one would wait for itself — with a cap of 1,
/// forever. The helper refuses at once (the second `acquire` below panics).
#[test]
#[should_panic(expected = "nested acquire")]
fn a_nested_acquire_on_one_thread_panics_instead_of_waiting() {
    let _serial = serial();
    let _first = spawn_cap::acquire();
    let _second = spawn_cap::acquire();
}

/// The cap is a whole number of at least 1; anything else is the default (4). The table pins
/// each edge: the parse, the `> 0` filter and the default.
#[test]
fn the_cap_text_is_a_whole_number_of_at_least_one_else_four() {
    assert_eq!(spawn_cap::DEFAULT_CAP, 4);
    for (text, want) in [
        (None, 4),
        (Some(""), 4),
        (Some("   "), 4),
        (Some("0"), 4),
        (Some("00"), 4),
        (Some("-3"), 4),
        (Some("2.5"), 4),
        (Some("abc"), 4),
        (Some("2x"), 4),
        (Some("99999999999999999999999999"), 4),
        (Some("1"), 1),
        (Some("2"), 2),
        (Some("4"), 4),
        (Some("7"), 7),
        (Some(" 3 "), 3),
        (Some("+5"), 5),
        (Some("16"), 16),
    ] {
        assert_eq!(spawn_cap::parse_cap(text), want, "{text:?}");
    }
}

/// The nextest bound (`.config/nextest.toml`) names every test file that takes permits —
/// under nextest each test is its own process and the in-process set bounds nothing, so a file
/// that takes permits but is not in the group is unbounded there.
#[test]
fn every_file_that_takes_permits_is_in_the_nextest_group() {
    let root = std::path::Path::new(env!("CARGO_MANIFEST_DIR"));
    let config = std::fs::read_to_string(root.join(".config/nextest.toml")).unwrap();
    assert!(
        config.contains("[test-groups]") && config.contains("spawning-moai"),
        "the spawning-moai test group is missing from .config/nextest.toml"
    );
    let mut users = Vec::new();
    for entry in std::fs::read_dir(root.join("tests")).unwrap() {
        let path = entry.unwrap().path();
        if path.extension().is_some_and(|e| e == "rs") {
            let text = std::fs::read_to_string(&path).unwrap();
            if text.contains("common/spawn_cap.rs") {
                users.push(path.file_stem().unwrap().to_string_lossy().into_owned());
            }
        }
    }
    users.retain(|n| n != "spawn_cap_test");
    users.sort();
    assert!(!users.is_empty(), "no test file takes permits any more");
    for name in &users {
        assert!(
            config.contains(&format!("binary({name})")),
            "tests/{name}.rs takes spawn permits but is not in the nextest group: {config}"
        );
    }
}
