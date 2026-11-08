//! Pure update coordination; does not discover hosts or perform updates.

#[derive(Debug, Clone, Copy, PartialEq, Eq)]
pub enum HostState {
    PendingPreflight,
    PreflightPassed,
    PreflightFailed,
    Ready,
    Applying,
    Succeeded,
    Failed,
    RecoveryPending,
    Recovered,
    RecoveryConflict,
}

#[derive(Debug, Clone, PartialEq, Eq)]
pub struct Coordinator {
    states: Vec<HostState>,
    applying: bool,
}

impl Coordinator {
    pub fn new(count: usize) -> Option<Self> {
        (count > 0).then(|| Self { states: vec![HostState::PendingPreflight; count], applying: false })
    }

    pub fn states(&self) -> &[HostState] { &self.states }

    pub fn preflight(&mut self, host: usize, passed: bool) -> Result<(), &'static str> {
        let state = self.states.get_mut(host).ok_or("unknown host")?;
        if self.applying || *state != HostState::PendingPreflight { return Err("invalid transition"); }
        *state = if passed { HostState::PreflightPassed } else { HostState::PreflightFailed };
        Ok(())
    }

    pub fn begin_apply(&mut self) -> Result<(), &'static str> {
        if self.applying || !self.states.iter().all(|state| *state == HostState::PreflightPassed) {
            return Err("preflight incomplete or failed");
        }
        self.applying = true;
        self.states.fill(HostState::Ready);
        Ok(())
    }

    pub fn start_host(&mut self, host: usize) -> Result<(), &'static str> {
        let state = self.states.get_mut(host).ok_or("unknown host")?;
        if !self.applying || *state != HostState::Ready { return Err("invalid transition"); }
        *state = HostState::Applying;
        Ok(())
    }

    pub fn applied(&mut self, host: usize, succeeded: bool) -> Result<(), &'static str> {
        let state = self.states.get_mut(host).ok_or("unknown host")?;
        if !self.applying || *state != HostState::Applying { return Err("invalid transition"); }
        *state = if succeeded { HostState::Succeeded } else { HostState::Failed };
        Ok(())
    }

    pub fn begin_recovery(&mut self, host: usize) -> Result<(), &'static str> {
        let state = self.states.get_mut(host).ok_or("unknown host")?;
        if *state != HostState::Failed { return Err("only failed hosts can recover"); }
        *state = HostState::RecoveryPending;
        Ok(())
    }

    pub fn recovered(&mut self, host: usize, conflict: bool) -> Result<(), &'static str> {
        let state = self.states.get_mut(host).ok_or("unknown host")?;
        if *state != HostState::RecoveryPending { return Err("invalid transition"); }
        *state = if conflict { HostState::RecoveryConflict } else { HostState::Recovered };
        Ok(())
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn cannot_apply_before_every_preflight() {
        assert_eq!(Coordinator::new(0), None);
        let mut c = Coordinator::new(2).unwrap();
        c.preflight(0, true).unwrap();
        assert!(c.begin_apply().is_err());
        c.preflight(1, false).unwrap();
        assert!(c.begin_apply().is_err());
        assert_eq!(c.states(), &[HostState::PreflightPassed, HostState::PreflightFailed]);
    }

    #[test]
    fn recovery_never_changes_successful_host() {
        let mut c = Coordinator::new(2).unwrap();
        c.preflight(0, true).unwrap(); c.preflight(1, true).unwrap();
        c.begin_apply().unwrap();
        c.start_host(0).unwrap(); c.applied(0, true).unwrap();
        assert_eq!(c.states()[1], HostState::Ready);
        c.start_host(1).unwrap(); c.applied(1, false).unwrap();
        assert!(c.begin_recovery(0).is_err());
        c.begin_recovery(1).unwrap(); c.recovered(1, true).unwrap();
        assert_eq!(c.states(), &[HostState::Succeeded, HostState::RecoveryConflict]);
    }

    #[test]
    fn recovery_status_and_invalid_transitions_remain_visible() {
        let mut c = Coordinator::new(1).unwrap();
        assert!(c.applied(0, true).is_err());
        assert!(c.preflight(3, true).is_err());
        c.preflight(0, true).unwrap(); c.begin_apply().unwrap();
        c.start_host(0).unwrap(); c.applied(0, false).unwrap(); c.begin_recovery(0).unwrap(); c.recovered(0, false).unwrap();
        assert_eq!(c.states(), &[HostState::Recovered]);
        assert!(c.applied(0, true).is_err());
    }
}
