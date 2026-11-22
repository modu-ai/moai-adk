//! The monthly cost estimate of the CI plan (SPEC-20261122-project-ci-design REQ-6): minutes
//! per run, rounded up to a whole minute, times runs per month, times the per-minute price from
//! a dated data file. A public repository shows zero for the standard runners. The result is an
//! estimate with a stated range; the tool never sets a budget, it only compares.
//!
//! Money is counted in mills (thousandths of a US dollar) so no sum is a float.

use serde::Deserialize;

/// The dated price table, embedded from `templates/ci/prices.json`.
const PRICES_JSON: &str = include_str!("../templates/ci/prices.json");

/// Per-minute prices in mills for the three standard hosted runners.
#[derive(Clone, Debug, Deserialize, PartialEq, Eq)]
pub struct Mills {
    pub linux: u32,
    pub windows: u32,
    pub macos: u32,
}

/// The price table and the day it was read.
#[derive(Clone, Debug, Deserialize, PartialEq, Eq)]
pub struct Prices {
    pub read_on: String,
    pub source: String,
    pub per_minute_mills: Mills,
}

/// The embedded price table.
pub fn prices() -> Prices {
    serde_json::from_str(PRICES_JSON).expect("templates/ci/prices.json is valid")
}

/// `0.062`-style dollars of a mills value.
pub fn dollars(mills: u32) -> String {
    format!("{}.{:03}", mills / 1000, mills % 1000)
}

/// Runs per month when the project has too few commits to count: the floor of the estimate.
pub const MIN_RUNS_PER_MONTH: u32 = 20;
/// Nightly runs per month (one a night).
pub const NIGHTLY_RUNS: u32 = 30;

/// One planned job's load: the runner, the seconds one run takes and the runs per month.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct JobLoad {
    /// `linux`, `windows` or `macos`.
    pub system: String,
    pub seconds_per_run: u32,
    pub runs_per_month: u32,
}

/// The estimate the preview prints.
#[derive(Clone, Debug, PartialEq, Eq)]
pub struct Estimate {
    /// Billed minutes per month, each run rounded up to a whole minute.
    pub minutes: u32,
    /// Estimated cost in mills; zero for a public repository's standard runners.
    pub mills: u32,
    /// The range around the estimate, in mills: 75 percent to 125 percent.
    pub low_mills: u32,
    pub high_mills: u32,
    /// A public repository: standard runners are free.
    pub free: bool,
    /// A private repository whose estimate exceeds the stated budget.
    pub over_budget: bool,
    /// The day the price table was read.
    pub prices_read_on: String,
}

/// Minutes one run takes by default, per project kind; the preview shows it and the user may
/// correct it.
pub fn default_minutes(kind: &str) -> u32 {
    match kind {
        "web" => 8,
        "api" => 6,
        "mobile" => 10,
        // `cli`, `library` and a kind this table does not know.
        _ => 5,
    }
}

/// The jobs the generated files hold, as loads: the Linux check on every run, a build-only job
/// per other named system (half the minutes, rounded up) when a build check is declared, and the
/// nightly run when the flow has one.
pub fn plan_jobs(
    answers: &crate::project_ci::CiAnswers,
    nightly: bool,
    has_build: bool,
    runs_per_month: u32,
) -> Vec<JobLoad> {
    let runs = runs_per_month.max(MIN_RUNS_PER_MONTH);
    let seconds = default_minutes(&answers.kind) * 60;
    let load = |system: &str, seconds_per_run: u32, runs_per_month: u32| JobLoad {
        system: system.into(),
        seconds_per_run,
        runs_per_month,
    };
    let mut jobs = vec![load("linux", seconds, runs)];
    if has_build {
        for system in ["macos", "windows"] {
            if answers.systems.iter().any(|s| s == system) {
                // A build-only job takes about half the time; the billed minute rounds up in
                // `estimate`.
                jobs.push(load(system, seconds / 2, runs));
            }
        }
    }
    if nightly {
        jobs.push(load("linux", seconds, NIGHTLY_RUNS));
    }
    jobs
}

/// The estimate of `jobs` for a repository with `answers`, from `prices`.
pub fn estimate(
    jobs: &[JobLoad],
    answers: &crate::project_ci::CiAnswers,
    prices: &Prices,
) -> Estimate {
    let free = answers.visibility == "public";
    let mut minutes = 0u32;
    let mut mills = 0u32;
    for job in jobs {
        let billed = job.seconds_per_run.div_ceil(60) * job.runs_per_month;
        let per_minute = match job.system.as_str() {
            "windows" => prices.per_minute_mills.windows,
            "macos" => prices.per_minute_mills.macos,
            _ => prices.per_minute_mills.linux,
        };
        minutes += billed;
        mills += billed * per_minute;
    }
    if free {
        mills = 0;
    }
    let over_budget = !free
        && answers
            .monthly_budget_usd
            .is_some_and(|usd| mills > usd * 1000);
    Estimate {
        minutes,
        mills,
        low_mills: mills / 4 * 3 + mills % 4 * 3 / 4,
        high_mills: mills / 4 * 5 + mills % 4 * 5 / 4,
        free,
        over_budget,
        prices_read_on: prices.read_on.clone(),
    }
}

#[cfg(test)]
mod tests {
    use super::*;
    use crate::project_ci::{CiAnswers, CiDeploy};

    fn answers(visibility: &str, systems: &[&str], budget: Option<u32>) -> CiAnswers {
        CiAnswers {
            kind: "cli".into(),
            systems: systems.iter().map(|s| s.to_string()).collect(),
            deploy: CiDeploy {
                enabled: false,
                target: None,
            },
            visibility: visibility.into(),
            monthly_budget_usd: budget,
            macos_price_ack: None,
            dependabot: None,
            strict: "report".into(),
        }
    }

    /// AC-3 (REQ-6): fixed inputs against a hand-computed value. A private CLI project (5 minutes
    /// per run) with Linux and Windows, the nightly run and 40 runs a month:
    /// Linux check 5 min x 40 x 6 = 1200, Windows build 3 min x 40 x 10 = 1200, nightly
    /// 5 min x 30 x 6 = 900, so 3300 mills = 3.30 dollars over 470 billed minutes.
    #[test]
    fn the_estimate_matches_a_hand_computed_value() {
        let a = answers("private", &["linux", "windows"], Some(4));
        let jobs = plan_jobs(&a, true, true, 40);
        assert_eq!(
            jobs,
            vec![
                JobLoad {
                    system: "linux".into(),
                    seconds_per_run: 300,
                    runs_per_month: 40
                },
                JobLoad {
                    system: "windows".into(),
                    seconds_per_run: 150,
                    runs_per_month: 40
                },
                JobLoad {
                    system: "linux".into(),
                    seconds_per_run: 300,
                    runs_per_month: 30
                },
            ]
        );
        let e = estimate(&jobs, &a, &prices());
        assert_eq!(e.minutes, 200 + 120 + 150);
        assert_eq!(e.mills, 3300);
        assert_eq!((e.low_mills, e.high_mills), (2475, 4125));
        assert!(!e.free && !e.over_budget);
        assert_eq!(e.prices_read_on, "2026-11-22");
    }

    /// N and N+1 of the budget line: equal is within the budget, one mill more is over; a public
    /// repository is free whatever the jobs are and never over a budget.
    #[test]
    fn the_budget_line_and_the_public_repository() {
        let jobs = plan_jobs(
            &answers("private", &["linux", "windows"], None),
            true,
            true,
            40,
        );
        assert!(
            !estimate(
                &jobs,
                &answers("private", &["linux", "windows"], Some(3300 / 1000 + 1)),
                &prices()
            )
            .over_budget
        );
        let exact = vec![JobLoad {
            system: "linux".into(),
            seconds_per_run: 60,
            runs_per_month: 1000,
        }];
        // 1 minute x 1000 runs x 6 mills = 6000 mills = 6 dollars exactly.
        assert_eq!(
            estimate(&exact, &answers("private", &["linux"], Some(6)), &prices()).mills,
            6000
        );
        assert!(!estimate(&exact, &answers("private", &["linux"], Some(6)), &prices()).over_budget);
        assert!(estimate(&exact, &answers("private", &["linux"], Some(5)), &prices()).over_budget);
        assert!(estimate(&exact, &answers("private", &["linux"], Some(0)), &prices()).over_budget);
        let none = estimate(&exact, &answers("private", &["linux"], None), &prices());
        assert!(!none.over_budget, "no stated budget, nothing to compare");
        let public = estimate(
            &jobs,
            &answers("public", &["linux", "windows"], None),
            &prices(),
        );
        assert!(public.free);
        assert_eq!(
            (public.mills, public.low_mills, public.high_mills),
            (0, 0, 0)
        );
        assert!(public.minutes > 0, "minutes are still counted");
        assert!(!public.over_budget);
    }

    /// Each run is rounded up to a whole minute before it is priced, and the runs have a floor.
    #[test]
    fn minutes_round_up_and_runs_have_a_floor() {
        let a = answers("private", &["linux"], None);
        let one = |seconds| {
            vec![JobLoad {
                system: "linux".into(),
                seconds_per_run: seconds,
                runs_per_month: 10,
            }]
        };
        assert_eq!(estimate(&one(60), &a, &prices()).minutes, 10);
        assert_eq!(estimate(&one(61), &a, &prices()).minutes, 20);
        assert_eq!(estimate(&one(1), &a, &prices()).minutes, 10);
        assert_eq!(estimate(&one(0), &a, &prices()).minutes, 0);
        let few = plan_jobs(&a, false, true, 3);
        assert_eq!(few[0].runs_per_month, MIN_RUNS_PER_MONTH);
        let enough = plan_jobs(&a, false, true, MIN_RUNS_PER_MONTH);
        assert_eq!(enough[0].runs_per_month, MIN_RUNS_PER_MONTH);
        let more = plan_jobs(&a, false, true, MIN_RUNS_PER_MONTH + 1);
        assert_eq!(more[0].runs_per_month, MIN_RUNS_PER_MONTH + 1);
    }

    /// An extra system adds a job only with a declared build check; without the nightly flow
    /// there is no nightly load; macOS is priced at its own rate.
    #[test]
    fn extra_systems_need_a_build_check_and_macos_has_its_own_price() {
        let a = answers("private", &["linux", "macos", "windows"], None);
        assert_eq!(plan_jobs(&a, false, false, 40).len(), 1);
        let jobs = plan_jobs(&a, false, true, 40);
        assert_eq!(
            jobs.iter().map(|j| j.system.as_str()).collect::<Vec<_>>(),
            ["linux", "macos", "windows"]
        );
        // macOS 3 min x 40 x 62 = 7440, Windows 1200, Linux 1200.
        assert_eq!(estimate(&jobs, &a, &prices()).mills, 1200 + 7440 + 1200);
    }

    #[test]
    fn default_minutes_follow_the_project_kind() {
        let want = [
            ("web", 8),
            ("api", 6),
            ("cli", 5),
            ("library", 5),
            ("mobile", 10),
        ];
        for (kind, minutes) in want {
            assert_eq!(default_minutes(kind), minutes, "{kind}");
        }
        for kind in crate::project_ci::KINDS {
            assert!(want.iter().any(|(k, _)| k == kind), "{kind} has no default");
        }
    }

    /// The embedded table carries the day it was read and the prices the SPEC recorded.
    #[test]
    fn the_embedded_price_table_is_dated_and_matches_the_spec() {
        let p = prices();
        assert_eq!(p.read_on, "2026-11-22");
        assert_eq!(
            p.per_minute_mills,
            Mills {
                linux: 6,
                windows: 10,
                macos: 62
            }
        );
        assert!(p.source.contains("docs.github.com"), "{}", p.source);
        assert_eq!(dollars(62), "0.062");
        assert_eq!(dollars(3300), "3.300");
        assert_eq!(dollars(0), "0.000");
    }
}
