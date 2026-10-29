package claim

import "time"

// lockPolicy is how a held lock is waited for. The zero value is the built-in
// polling: lockTries attempts, lockWait apart, on the real clock.
type lockPolicy struct {
	tries int
	wait  time.Duration
	sleep func(time.Duration)
}

func (p lockPolicy) resolve() (int, time.Duration, func(time.Duration)) {
	tries, wait, sleep := p.tries, p.wait, p.sleep
	if tries == 0 {
		tries = lockTries
	}
	if wait == 0 {
		wait = lockWait
	}
	if sleep == nil {
		sleep = time.Sleep
	}
	return tries, wait, sleep
}

// LockWaitBudget is how long the built-in wait polls for a held lock before it
// answers ErrBusy, at the nominal polling interval. Sleeps can only run longer than
// asked, so a wait never ends sooner than this; how much longer depends on the load.
func LockWaitBudget() time.Duration { return time.Duration(lockTries) * lockWait }
