    /// Audit B2 and N9: one long word without a space, of the note limit's size, is screened in
    /// time that grows with its length, not with its square. Four shapes of word and one shape at
    /// the end of a long word run at once on their own threads, so the test takes about the time of
    /// the slowest one; a square-time regression (40 KB took 30 s) misses the 15 s bound and fails
    /// there instead of holding the test for minutes (the stuck threads die with the test process).
    #[test]
    fn secret_screen_time_on_one_long_word_has_a_bound() {
        use std::time::{Duration, Instant};
        let size = crate::ledger::workflow::MAX_NOTE_BYTES;
        let key = j(&["AK", "IA", "ABCDEFGHIJKLMNOP"]);
        let units = ["a,", "a:", "a=", "{\"a\":1,\"b\":[2,3]},"];
        let plain = |unit: &str| unit.repeat(size / unit.len() + 1)[..size].to_string();
        let mut jobs: Vec<(String, String, usize)> = units
            .iter()
            .map(|unit| (format!("{unit} x {size}"), plain(unit), 0))
            .collect();
        let tail = format!("{},{key}", &plain("a,")[..size - key.len() - 1]);
        jobs.push((format!("a, x {size} with a shape at the end"), tail, 1));
        let (tx, rx) = std::sync::mpsc::channel();
        for (at, (_, text, _)) in jobs.iter().enumerate() {
            let (tx, text) = (tx.clone(), text.clone());
            std::thread::spawn(move || {
                let _ = tx.send((at, secret_words(&text)));
            });
        }
        let deadline = Instant::now() + Duration::from_secs(15);
        for _ in 0..jobs.len() {
            let wait = deadline.saturating_duration_since(Instant::now());
            let (at, words) = rx
                .recv_timeout(wait)
                .unwrap_or_else(|_| panic!("a screen of {size} bytes took more than 15 s"));
            assert_eq!(words, jobs[at].2, "{}", jobs[at].0);
        }
    }
