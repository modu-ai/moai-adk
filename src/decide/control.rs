//! The control gate of `moai decide` (SPEC-20261118-decide-control): the decision model is off
//! until the settings say it is on, and it may be called a bounded number of times a day.
//!
//! The gate is pure. It takes the settings, the day's usage, today's date, the request text and
//! the sender as arguments, and it opens no file, reads no clock, holds no key and no address.
//! Reading and writing the settings and the usage, and today's date, belong to the command
//! that calls it; the sender is [`super::transport`] or a test double.
//!
//! The order is fixed: off refuses first (before the request is read), then the day, then the
//! day's limit, then the request checks of [`super::prepare`], and only then is the call
//! counted and the sender called. A call that fails after the count still counts (D-2).
//! Every refusal means the same thing to the caller: the model does not decide, the leader
//! does (`fallback: leader`).

use super::outbound::Who;
use super::provider::Provider;
use super::{conclude, prepare, Prepared, Record, Refusal};
use serde::Deserialize;

/// Calls a day when the settings name no limit.
pub const CALLS_PER_DAY_DEFAULT: u32 = 20;
/// Largest limit the settings may name.
pub const CALLS_PER_DAY_MAX: u32 = 1000;

/// What the settings say. The default is off.
#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct Settings {
    pub enabled: bool,
    pub provider: Provider,
    pub calls_per_day: u32,
}

impl Default for Settings {
    fn default() -> Self {
        Settings {
            enabled: false,
            provider: Provider::Typesafe,
            calls_per_day: CALLS_PER_DAY_DEFAULT,
        }
    }
}

#[derive(Deserialize)]
#[serde(deny_unknown_fields)]
struct RawSettings {
    enabled: Option<bool>,
    provider: Option<String>,
    calls_per_day: Option<u32>,
}

impl Settings {
    /// The settings of a JSON text; every field is optional and the missing ones keep the
    /// default (so `{}` is off). Anything else refuses with `settings_invalid`.
    pub fn from_json(text: &str) -> Result<Settings, Refusal> {
        let invalid = |why: &str| Refusal::new("settings_invalid", format!("the settings {why}"));
        let raw: RawSettings = serde_json::from_str(text)
            .map_err(|_| invalid("are not a JSON object of known fields"))?;
        let mut settings = Settings::default();
        if let Some(enabled) = raw.enabled {
            settings.enabled = enabled;
        }
        if let Some(name) = raw.provider {
            settings.provider = Provider::from_name(&name)
                .ok_or_else(|| invalid("name a provider that is not compiled in"))?;
        }
        if let Some(limit) = raw.calls_per_day {
            if !(1..=CALLS_PER_DAY_MAX).contains(&limit) {
                return Err(invalid("name a daily limit outside 1 to 1000"));
            }
            settings.calls_per_day = limit;
        }
        Ok(settings)
    }
}

/// Whether `day` is `YYYY-MM-DD` with a month 01 to 12 and a day 01 to 31, ASCII digits only.
fn is_day(day: &str) -> bool {
    let b = day.as_bytes();
    if b.len() != 10 || b[4] != b'-' || b[7] != b'-' {
        return false;
    }
    let digits = |range: std::ops::Range<usize>| -> Option<u32> {
        let part = &b[range];
        if part.iter().all(u8::is_ascii_digit) {
            std::str::from_utf8(part).ok()?.parse().ok()
        } else {
            None
        }
    };
    matches!(
        (digits(0..4), digits(5..7), digits(8..10)),
        (Some(_), Some(1..=12), Some(1..=31))
    )
}

/// The calls counted for one day: `day` is `YYYY-MM-DD`, empty before the first call.
#[derive(Debug, Clone, Default, PartialEq, Eq)]
pub struct Usage {
    pub day: String,
    pub calls: u32,
}

/// Put `text` to the decision model once, if the settings and the day's limit allow it.
pub fn decide(
    settings: &Settings,
    usage: &mut Usage,
    today: &str,
    text: &str,
    who: &Who,
    send: impl FnOnce(&Prepared) -> Result<Vec<u8>, Refusal>,
) -> Result<Record, Refusal> {
    if !settings.enabled {
        return Err(Refusal::new("decide_off", "the decision model is off"));
    }
    if !is_day(today) {
        return Err(Refusal::new(
            "clock_invalid",
            "the day is not a date of the form YYYY-MM-DD",
        ));
    }
    // A later day counts again from zero; the same or an earlier day (a clock set back) keeps
    // the stored count. `YYYY-MM-DD` orders as text.
    if today > usage.day.as_str() {
        usage.day = today.to_string();
        usage.calls = 0;
    }
    if usage.calls >= settings.calls_per_day {
        return Err(Refusal::new(
            "decide_limit",
            "the decision model's calls for the day are used up",
        ));
    }
    let prepared = prepare(text, settings.provider, who)?;
    // Counted before the call: a slow or failed call still uses the day's limit (D-2).
    usage.calls += 1;
    let response = send(&prepared)?;
    conclude(&prepared, &response)
}

#[cfg(test)]
mod tests {
    use super::super::provider::{OPENAI_ENDPOINT, TYPESAFE_ENDPOINT};
    use super::super::testing::{request_json, who};
    use super::*;
    use serde_json::json;
    use std::cell::{Cell, RefCell};

    const TYPESAFE_ANSWER: &str = r#"{"model":"jev-test-1","answers":{"decision":{"type":"choice","choice":"nextest","confidence":0.9,"probabilities":{"nextest":0.8,"cargo-test":0.2}}}}"#;
    const OPENAI_ANSWER: &str = r#"{"answers":[{"type":"choice","name":"decision","choice":"nextest","confidence":0.8,"probabilities":[{"value":"nextest","probability":0.7},{"value":"cargo-test","probability":0.3}]}]}"#;

    fn on(limit: u32) -> Settings {
        Settings {
            enabled: true,
            provider: Provider::Typesafe,
            calls_per_day: limit,
        }
    }

    /// A sender that counts its calls and answers like the Jev server.
    struct Counting {
        sends: Cell<u32>,
    }

    impl Counting {
        fn new() -> Self {
            Counting {
                sends: Cell::new(0),
            }
        }
        fn send(&self, _: &Prepared) -> Result<Vec<u8>, Refusal> {
            self.sends.set(self.sends.get() + 1);
            Ok(TYPESAFE_ANSWER.as_bytes().to_vec())
        }
    }

    fn code(result: Result<Record, Refusal>) -> &'static str {
        result.expect_err("a refusal").code
    }

    #[test]
    fn the_default_is_off_and_off_calls_nothing() {
        assert!(!Settings::default().enabled);
        assert_eq!(Settings::from_json("{}").unwrap(), Settings::default());
        assert_eq!(
            Settings::from_json(r#"{"enabled": false}"#).unwrap(),
            Settings::default()
        );
        let sender = Counting::new();
        let mut usage = Usage::default();
        // A request that would pass, one that would not, and a day that is no date: off comes
        // first and reads none of them.
        for (text, day) in [
            (request_json(|_| {}), "2026-11-18"),
            ("this is not json".to_string(), "2026-11-18"),
            (request_json(|_| {}), "yesterday"),
        ] {
            let refusal = decide(&Settings::default(), &mut usage, day, &text, &who(), |p| {
                sender.send(p)
            })
            .expect_err("off refuses");
            assert_eq!(refusal.code, "decide_off");
            assert_eq!(refusal.to_json()["fallback"], json!("leader"));
        }
        assert_eq!(sender.sends.get(), 0);
        assert_eq!(usage, Usage::default());
    }

    #[test]
    fn the_limit_stops_the_call_after_n_and_a_later_day_starts_again() {
        let sender = Counting::new();
        let mut usage = Usage::default();
        let settings = on(3);
        let request = request_json(|_| {});
        for n in 1..=3 {
            let record = decide(&settings, &mut usage, "2026-11-18", &request, &who(), |p| {
                sender.send(p)
            })
            .expect("a call under the limit");
            assert_eq!(record.chosen, "nextest");
            assert_eq!((sender.sends.get(), usage.calls), (n, n));
        }
        // Call N+1 is refused and nothing moves.
        let fourth = decide(&settings, &mut usage, "2026-11-18", &request, &who(), |p| {
            sender.send(p)
        });
        assert_eq!(code(fourth), "decide_limit");
        assert_eq!((sender.sends.get(), usage.calls), (3, 3));
        // A later day counts again from one.
        decide(&settings, &mut usage, "2026-11-19", &request, &who(), |p| {
            sender.send(p)
        })
        .expect("a new day");
        assert_eq!(
            usage,
            Usage {
                day: "2026-11-19".into(),
                calls: 1
            }
        );
        // A clock set back does not start a new count: the stored day stands.
        usage.calls = 3;
        let back = decide(&settings, &mut usage, "2026-11-18", &request, &who(), |p| {
            sender.send(p)
        });
        assert_eq!(code(back), "decide_limit");
        assert_eq!(
            usage,
            Usage {
                day: "2026-11-19".into(),
                calls: 3
            }
        );
        assert_eq!(sender.sends.get(), 4);
        // The limit is checked before the request is read.
        let unread = decide(
            &settings,
            &mut usage,
            "2026-11-19",
            "not json",
            &who(),
            |p| sender.send(p),
        );
        assert_eq!(code(unread), "decide_limit");
    }

    #[test]
    fn a_failed_call_still_counts_and_the_provider_comes_from_the_settings() {
        let mut usage = Usage::default();
        let request = request_json(|_| {});
        let failed = decide(&on(2), &mut usage, "2026-11-18", &request, &who(), |_| {
            Err(Refusal::new("transport_timeout", "the server was too slow"))
        });
        assert_eq!(code(failed), "transport_timeout");
        assert_eq!(usage.calls, 1);

        for (provider, endpoint, answer, prefix) in [
            (
                Provider::Typesafe,
                TYPESAFE_ENDPOINT,
                TYPESAFE_ANSWER,
                "model:typesafe/",
            ),
            (
                Provider::Openai,
                OPENAI_ENDPOINT,
                OPENAI_ANSWER,
                "model:openai/",
            ),
        ] {
            let seen = RefCell::new(String::new());
            let settings = Settings { provider, ..on(5) };
            let record = decide(&settings, &mut usage, "2026-11-18", &request, &who(), |p| {
                *seen.borrow_mut() = p.endpoint().to_string();
                Ok(answer.as_bytes().to_vec())
            })
            .expect("a call");
            assert_eq!(*seen.borrow(), endpoint);
            assert!(
                record.decided_by.starts_with(prefix),
                "{}",
                record.decided_by
            );
        }
        assert_eq!(usage.calls, 3);
    }

    #[test]
    fn a_request_the_gate_refuses_is_neither_sent_nor_counted() {
        let sender = Counting::new();
        let mut usage = Usage {
            day: "2026-11-18".into(),
            calls: 1,
        };
        let key = format!("sk-proj-{}", "A1b2C3d4E5f6G7h8I9j0");
        let requests = [
            request_json(|v| v["reversible"] = json!(false)),
            request_json(|v| v["category"] = json!("money")),
            request_json(|v| v["evidence"] = json!({"log": format!("the token is {key}")})),
            "this is not json".to_string(),
        ];
        for request in &requests {
            let result = decide(&on(5), &mut usage, "2026-11-18", request, &who(), |p| {
                sender.send(p)
            });
            let refusal = result.expect_err("the request is refused");
            assert!(
                !["decide_off", "decide_limit"].contains(&refusal.code),
                "{}",
                refusal.code
            );
            assert_eq!(refusal.to_json()["fallback"], json!("leader"));
            assert!(!refusal.reason.contains(&key));
        }
        assert_eq!(sender.sends.get(), 0);
        assert_eq!(
            usage,
            Usage {
                day: "2026-11-18".into(),
                calls: 1
            }
        );
    }

    #[test]
    fn the_settings_text_is_read_strictly() {
        let bad = [
            "",
            "not json",
            "[]",
            "null",
            r#"{"enabled": "true"}"#,
            r#"{"enabled": 1}"#,
            r#"{"enabled": null, "extra": true}"#,
            r#"{"unknown": 1}"#,
            r#"{"provider": "acme"}"#,
            r#"{"provider": 7}"#,
            r#"{"provider": "Typesafe"}"#,
            r#"{"calls_per_day": 0}"#,
            r#"{"calls_per_day": 1001}"#,
            r#"{"calls_per_day": -1}"#,
            r#"{"calls_per_day": 1.5}"#,
            r#"{"calls_per_day": "5"}"#,
            r#"{"enabled": true,}"#,
        ];
        for text in bad {
            let refusal = Settings::from_json(text).expect_err(text);
            assert_eq!(refusal.code, "settings_invalid", "{text}");
            assert_eq!(refusal.to_json()["fallback"], json!("leader"));
        }
        assert_eq!(
            Settings::from_json(r#"{"enabled": true}"#).unwrap(),
            Settings {
                enabled: true,
                ..Settings::default()
            }
        );
        assert_eq!(
            Settings::from_json(
                r#"{"enabled": true, "provider": "openai", "calls_per_day": 1000}"#
            )
            .unwrap(),
            Settings {
                enabled: true,
                provider: Provider::Openai,
                calls_per_day: 1000
            }
        );
        assert_eq!(
            Settings::from_json(r#"{"calls_per_day": 1}"#)
                .unwrap()
                .calls_per_day,
            1
        );
    }

    #[test]
    fn a_day_that_is_not_a_date_is_refused() {
        let sender = Counting::new();
        let mut usage = Usage::default();
        let request = request_json(|_| {});
        for day in [
            "",
            "2026-1-18",
            "2026-13-01",
            "2026-00-10",
            "2026-11-00",
            "2026-11-32",
            "20261118",
            "2026/11/18",
            "2026-11-18 ",
            "2026-11-18T00:00:00Z",
            "\u{ff12}\u{ff10}\u{ff12}\u{ff16}-11-18",
            "abcd-ef-gh",
        ] {
            let result = decide(&on(5), &mut usage, day, &request, &who(), |p| {
                sender.send(p)
            });
            assert_eq!(code(result), "clock_invalid", "{day:?}");
        }
        assert_eq!(sender.sends.get(), 0);
        assert_eq!(usage, Usage::default());
        // The edges of a date that is accepted.
        for day in ["2026-01-01", "2026-12-31"] {
            decide(&on(5), &mut usage, day, &request, &who(), |p| {
                sender.send(p)
            })
            .expect(day);
        }
    }
}
