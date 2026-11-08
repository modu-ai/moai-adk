//! Pure policy for the GitHub adapter: Link pagination, rate-limit backoff, time arithmetic.
//! Sources (checked 2026-11-08): GitHub REST best practices, rate limits and pagination pages
//! cited in Issue #2008.
use super::transport::HttpResponse;

/// Wall clock and sleep, injected so tests never sleep.
pub trait Clock {
    fn now(&self) -> i64;
    fn sleep(&self, secs: u64);
}

pub struct SystemClock;

impl Clock for SystemClock {
    fn now(&self) -> i64 {
        std::time::SystemTime::now()
            .duration_since(std::time::UNIX_EPOCH)
            .map(|d| d.as_secs() as i64)
            .unwrap_or(0)
    }
    fn sleep(&self, secs: u64) {
        std::thread::sleep(std::time::Duration::from_secs(secs));
    }
}

/// Returns the `rel="next"` URL of a `Link` header exactly as GitHub wrote it (REQ-FORGE-005).
pub fn next_link(link: Option<&str>) -> Option<String> {
    link?.split(',').find_map(|part| {
        let (url, params) = part.trim().split_once(';')?;
        let url = url.trim().strip_prefix('<')?.strip_suffix('>')?;
        params
            .split(';')
            .any(|p| p.trim() == "rel=\"next\"")
            .then(|| url.to_string())
    })
}

/// What to do with a response before using it.
#[derive(Debug, PartialEq, Eq)]
pub enum Verdict {
    Use,
    /// Wait this many seconds, then resend the same request.
    Retry(u64),
    /// Authentication or authorization failure: stop, report the reason only.
    Blocked(&'static str),
}

/// REQ-FORGE-003: `retry-after` seconds first; then `x-ratelimit-remaining: 0` waits until
/// `x-ratelimit-reset`; any other 403/429 is a secondary limit and waits at least 60 s with
/// exponential growth per attempt. 401 and other 403s are blocked.
pub fn judge(resp: &HttpResponse, attempt: u32, now: i64) -> Verdict {
    let num = |name: &str| resp.header(name).and_then(|v| v.trim().parse::<i64>().ok());
    if resp.status == 401 {
        return Verdict::Blocked("authentication failed (401)");
    }
    if !(resp.status == 403 || resp.status == 429) {
        return Verdict::Use;
    }
    if let Some(secs) = num("retry-after") {
        return Verdict::Retry(secs.max(0) as u64);
    }
    if num("x-ratelimit-remaining") == Some(0) {
        if let Some(reset) = num("x-ratelimit-reset") {
            return Verdict::Retry((reset - now).max(0) as u64);
        }
    }
    let rate_message = resp
        .body
        .as_ref()
        .and_then(|b| b["message"].as_str())
        .map(|m| m.to_ascii_lowercase().contains("rate limit"))
        .unwrap_or(false);
    if resp.status == 429 || rate_message {
        return Verdict::Retry(60u64.saturating_mul(1u64 << attempt.min(10)));
    }
    Verdict::Blocked("access denied (403)")
}

/// Days-from-civil conversion (proleptic Gregorian, no external crate).
fn days_from_civil(y: i64, m: i64, d: i64) -> i64 {
    let y = if m <= 2 { y - 1 } else { y };
    let era = y.div_euclid(400);
    let yoe = y - era * 400;
    let mp = (m + 9) % 12;
    let doy = (153 * mp + 2) / 5 + d - 1;
    let doe = yoe * 365 + yoe / 4 - yoe / 100 + doy;
    era * 146_097 + doe - 719_468
}

/// Parses `YYYY-MM-DDTHH:MM:SSZ` (the form GitHub returns) into Unix seconds.
pub fn parse_utc(s: &str) -> Option<i64> {
    let b = s.as_bytes();
    if b.len() != 20 || b[4] != b'-' || b[7] != b'-' || b[10] != b'T' || b[19] != b'Z' {
        return None;
    }
    let n = |r: std::ops::Range<usize>| s.get(r)?.parse::<i64>().ok();
    let (y, mo, d, h, mi, se) = (
        n(0..4)?,
        n(5..7)?,
        n(8..10)?,
        n(11..13)?,
        n(14..16)?,
        n(17..19)?,
    );
    if !(1..=12).contains(&mo) || !(1..=31).contains(&d) || h > 23 || mi > 59 || se > 60 {
        return None;
    }
    Some(days_from_civil(y, mo, d) * 86_400 + h * 3600 + mi * 60 + se)
}

/// Formats Unix seconds as `YYYY-MM-DDTHH:MM:SSZ`.
pub fn format_utc(secs: i64) -> String {
    let (days, rem) = (secs.div_euclid(86_400), secs.rem_euclid(86_400));
    let z = days + 719_468;
    let era = z.div_euclid(146_097);
    let doe = z - era * 146_097;
    let yoe = (doe - doe / 1460 + doe / 36_524 - doe / 146_096) / 365;
    let doy = doe - (365 * yoe + yoe / 4 - yoe / 100);
    let mp = (5 * doy + 2) / 153;
    let d = doy - (153 * mp + 2) / 5 + 1;
    let m = if mp < 10 { mp + 3 } else { mp - 9 };
    let y = yoe + era * 400 + i64::from(m <= 2);
    format!(
        "{y:04}-{m:02}-{d:02}T{:02}:{:02}:{:02}Z",
        rem / 3600,
        rem % 3600 / 60,
        rem % 60
    )
}

#[cfg(test)]
mod tests {
    use super::*;
    use std::collections::BTreeMap;

    fn resp(status: u16, headers: &[(&str, &str)]) -> HttpResponse {
        HttpResponse {
            status,
            headers: headers
                .iter()
                .map(|(k, v)| (k.to_string(), v.to_string()))
                .collect::<BTreeMap<_, _>>(),
            body: None,
        }
    }

    #[test]
    fn link_next_is_taken_verbatim() {
        let h = r#"<https://api.github.com/repositories/1/issues?page=2&per_page=100>; rel="next", <https://api.github.com/repositories/1/issues?page=5>; rel="last""#;
        assert_eq!(
            next_link(Some(h)).as_deref(),
            Some("https://api.github.com/repositories/1/issues?page=2&per_page=100")
        );
        assert_eq!(next_link(Some(r#"<https://x/>; rel="prev""#)), None);
        assert_eq!(next_link(None), None);
    }

    #[test]
    fn utc_round_trip() {
        for s in [
            "1970-01-01T00:00:00Z",
            "2026-11-08T09:48:15Z",
            "2024-02-29T23:59:59Z",
        ] {
            assert_eq!(format_utc(parse_utc(s).unwrap()), s);
        }
        assert_eq!(parse_utc("2026-11-08T09:48:15Z"), Some(1_794_131_295));
        assert_eq!(parse_utc("2026-11-08 09:48:15"), None);
    }

    #[test]
    fn backoff_order() {
        assert_eq!(judge(&resp(200, &[]), 0, 0), Verdict::Use);
        assert_eq!(judge(&resp(304, &[]), 0, 0), Verdict::Use);
        assert_eq!(
            judge(&resp(403, &[("retry-after", "3")]), 0, 0),
            Verdict::Retry(3)
        );
        assert_eq!(
            judge(
                &resp(
                    403,
                    &[
                        ("x-ratelimit-remaining", "0"),
                        ("x-ratelimit-reset", "1100")
                    ]
                ),
                0,
                1000
            ),
            Verdict::Retry(100)
        );
        assert_eq!(judge(&resp(429, &[]), 0, 0), Verdict::Retry(60));
        assert_eq!(judge(&resp(429, &[]), 2, 0), Verdict::Retry(240));
        assert!(matches!(judge(&resp(401, &[]), 0, 0), Verdict::Blocked(_)));
        assert!(matches!(judge(&resp(403, &[]), 0, 0), Verdict::Blocked(_)));
    }
}
