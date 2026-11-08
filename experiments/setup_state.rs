//! SETUP state contract experiment; it performs no discovery or file writes.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum Phase { Inspecting, Interviewing, Planned, Applying, Validating, Ready, AwaitingInput, Blocked, Partial, NeedsReplan }

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub struct HostEvidence {
    pub files_created: bool,
    pub host_loaded: bool,
    pub execution_verified: bool,
}

#[derive(Debug, PartialEq, Eq)]
pub struct Setup {
    phase: Phase,
    hosts: Vec<HostEvidence>,
}

impl Setup {
    pub fn new(hosts: usize) -> Option<Self> {
        (hosts > 0).then(|| Self { phase: Phase::Inspecting, hosts: vec![HostEvidence { files_created: false, host_loaded: false, execution_verified: false }; hosts] })
    }
    pub fn phase(&self) -> Phase { self.phase }
    pub fn advance(&mut self, next: Phase) -> Result<(), &'static str> {
        let allowed = matches!((self.phase, next),
            (Phase::Inspecting, Phase::Interviewing) | (Phase::Interviewing, Phase::Planned) |
            (Phase::Planned, Phase::Applying) | (Phase::Applying, Phase::Validating));
        if !allowed { return Err("invalid transition"); }
        self.phase = next;
        Ok(())
    }
    pub fn record(&mut self, index: usize, evidence: HostEvidence) -> Result<(), &'static str> {
        if self.phase != Phase::Validating { return Err("validation not active"); }
        if evidence.host_loaded && !evidence.files_created || evidence.execution_verified && !evidence.host_loaded {
            return Err("inconsistent evidence");
        }
        let host = self.hosts.get_mut(index).ok_or("unknown host")?;
        *host = evidence;
        Ok(())
    }
    pub fn finish(&mut self) -> Result<(), &'static str> {
        if self.phase != Phase::Validating || !self.hosts.iter().all(|h| h.files_created && h.host_loaded && h.execution_verified) {
            return Err("host evidence incomplete");
        }
        self.phase = Phase::Ready;
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    fn validating(count: usize) -> Setup {
        let mut s = Setup::new(count).unwrap();
        for p in [Phase::Interviewing, Phase::Planned, Phase::Applying, Phase::Validating] { s.advance(p).unwrap(); }
        s
    }
    #[test]
    fn no_hosts_or_skipped_interview_are_ready() {
        assert_eq!(Setup::new(0), None);
        let mut s = Setup::new(2).unwrap();
        assert!(s.advance(Phase::Planned).is_err());
        assert!(s.finish().is_err());
    }
    #[test]
    fn files_and_one_host_are_not_both_host_success() {
        let mut s = validating(2);
        s.record(0, HostEvidence { files_created: true, host_loaded: true, execution_verified: true }).unwrap();
        s.record(1, HostEvidence { files_created: true, host_loaded: false, execution_verified: false }).unwrap();
        assert!(s.finish().is_err());
        assert_eq!(s.phase(), Phase::Validating);
    }
    #[test]
    fn all_hosts_need_observed_execution_and_ordered_evidence() {
        let mut s = validating(2);
        assert!(s.record(0, HostEvidence { files_created: true, host_loaded: false, execution_verified: true }).is_err());
        let complete = HostEvidence { files_created: true, host_loaded: true, execution_verified: true };
        s.record(0, complete).unwrap(); s.record(1, complete).unwrap();
        s.finish().unwrap();
        assert_eq!(s.phase(), Phase::Ready);
    }
}
