//! Text forms the Go build prints: `time.Duration` strings and its time stamps.

use std::time::{Duration, SystemTime, UNIX_EPOCH};

/// Go's `time.Duration.String()`: `0s`, `122ms`, `1.5s`, `1m30s`, `2m0s`, `1h0m0s`.
pub fn go_duration(d: Duration) -> String {
    let ns = d.as_nanos();
    if ns == 0 {
        return "0s".into();
    }
    if ns < 1_000 {
        return format!("{ns}ns");
    }
    if ns < 1_000_000 {
        return with_fraction(ns, 1_000, "µs", 3);
    }
    if ns < 1_000_000_000 {
        return with_fraction(ns, 1_000_000, "ms", 6);
    }
    let total = ns / 1_000_000_000;
    let rem = ns % 1_000_000_000;
    let (h, m, s) = (total / 3600, (total % 3600) / 60, total % 60);
    let secs = if rem == 0 {
        s.to_string()
    } else {
        let frac = format!("{rem:09}");
        format!("{s}.{}", frac.trim_end_matches('0'))
    };
    if h > 0 {
        format!("{h}h{m}m{secs}s")
    } else if m > 0 {
        format!("{m}m{secs}s")
    } else {
        format!("{secs}s")
    }
}

fn with_fraction(ns: u128, unit: u128, suffix: &str, width: usize) -> String {
    let (whole, rem) = (ns / unit, ns % unit);
    if rem == 0 {
        return format!("{whole}{suffix}");
    }
    let frac = format!("{rem:0width$}");
    format!("{whole}.{}{suffix}", frac.trim_end_matches('0'))
}

/// `Duration.Round(time.Millisecond)`: half rounds away from zero.
pub fn round_ms(d: Duration) -> Duration {
    let ns = d.as_nanos();
    let ms = (ns + 500_000) / 1_000_000;
    Duration::from_millis(ms as u64)
}

/// Go's `2006-01-02T15:04:05.000Z07:00` in the machine's local zone.
pub fn local_stamp(t: SystemTime) -> String {
    let since = t.duration_since(UNIX_EPOCH).unwrap_or_default();
    let secs = since.as_secs() as libc::time_t;
    let mut tm: libc::tm = unsafe { std::mem::zeroed() };
    unsafe { libc::localtime_r(&secs, &mut tm) };
    let off = tm.tm_gmtoff;
    let zone = if off == 0 {
        "Z".to_string()
    } else {
        let sign = if off < 0 { '-' } else { '+' };
        let a = off.abs();
        format!("{sign}{:02}:{:02}", a / 3600, (a % 3600) / 60)
    };
    format!(
        "{:04}-{:02}-{:02}T{:02}:{:02}:{:02}.{:03}{zone}",
        tm.tm_year + 1900,
        tm.tm_mon + 1,
        tm.tm_mday,
        tm.tm_hour,
        tm.tm_min,
        tm.tm_sec,
        since.subsec_millis()
    )
}

/// `2026-11-20T00:08:30Z` (RFC 3339, UTC, whole seconds).
pub fn utc_rfc3339(t: SystemTime) -> String {
    let secs = t.duration_since(UNIX_EPOCH).unwrap_or_default().as_secs() as libc::time_t;
    let mut tm: libc::tm = unsafe { std::mem::zeroed() };
    unsafe { libc::gmtime_r(&secs, &mut tm) };
    format!(
        "{:04}-{:02}-{:02}T{:02}:{:02}:{:02}Z",
        tm.tm_year + 1900,
        tm.tm_mon + 1,
        tm.tm_mday,
        tm.tm_hour,
        tm.tm_min,
        tm.tm_sec
    )
}
