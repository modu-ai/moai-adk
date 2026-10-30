package ledger

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"
)

// expiredSQL is an instant that has passed, for a lock that was released or whose holder
// is dead. The row stays so the lock's epoch goes on counting.
const expiredSQL = `'1970-01-01T00:00:00.000Z'`

// AcquireLock takes the named lock for ttl and returns its epoch. ErrLockHeld when
// another live holder has it, or the same holder asks again while its lock is live; a
// lock whose ttl ran out, or that was released or expired by Reap, is taken over with a
// higher epoch.
func (s *Store) AcquireLock(ctx context.Context, name string, holder LockHolder, ttl time.Duration) (int64, error) {
	if strings.TrimSpace(name) == "" || strings.TrimSpace(holder.SessionID) == "" || ttl <= 0 {
		return 0, fmt.Errorf("%w: a lock needs a name, a holder and a ttl", ErrInvalid)
	}
	var pid, start any
	if holder.PID > 0 {
		pid = holder.PID
	}
	start = nullIfEmpty(holder.PIDStart)
	var epoch int64
	err := s.write(ctx, func(tx *sql.Tx) error {
		var held int64
		var live bool
		switch err := tx.QueryRowContext(ctx, `SELECT epoch, expires_at > `+nowSQL+` FROM locks WHERE name=?`, name).Scan(&held, &live); {
		case err == sql.ErrNoRows:
			epoch = 1
			_, err := tx.ExecContext(ctx, `INSERT INTO locks(name,session_id,epoch,pid,pid_start,expires_at) VALUES(?,?,?,?,?,`+offsetSQL(ttl)+`)`, name, holder.SessionID, epoch, pid, start)
			if err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
			return nil
		case err != nil:
			return fmt.Errorf("ledger: %w", err)
		case live:
			return fmt.Errorf("%w: %s", ErrLockHeld, name)
		}
		epoch = held + 1
		_, err := tx.ExecContext(ctx, `UPDATE locks SET session_id=?, epoch=?, pid=?, pid_start=?, expires_at=`+offsetSQL(ttl)+` WHERE name=?`, holder.SessionID, epoch, pid, start, name)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		return nil
	})
	if err != nil {
		return 0, err
	}
	return epoch, nil
}

// ReleaseLock gives the lock up: its expiry is set in the past and the row stays, so the
// next AcquireLock has a higher epoch. ErrStaleEpoch when the epoch is not the lock's, or
// the lock was already released (or expired by Reap). A lock whose ttl merely ran out is
// still its holder's to release at its own epoch until someone takes it over: the epoch,
// not the clock, says whose it is.
func (s *Store) ReleaseLock(ctx context.Context, name string, epoch int64) error {
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE locks SET expires_at=`+expiredSQL+` WHERE name=? AND epoch=? AND expires_at<>`+expiredSQL, name, epoch)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("%w: lock %s is not held at epoch %d", ErrStaleEpoch, name, epoch)
		}
		return nil
	})
}

// RenewLock extends a live lock held at epoch to ttl from now. ErrStaleEpoch when the
// lock is not live at that epoch: one that ran out is taken again, not renewed.
func (s *Store) RenewLock(ctx context.Context, name string, epoch int64, ttl time.Duration) error {
	if ttl <= 0 {
		return fmt.Errorf("%w: a renewal needs a ttl", ErrInvalid)
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		res, err := tx.ExecContext(ctx, `UPDATE locks SET expires_at=`+offsetSQL(ttl)+` WHERE name=? AND epoch=? AND expires_at > `+nowSQL, name, epoch)
		if err != nil {
			return fmt.Errorf("ledger: %w", err)
		}
		if n, _ := res.RowsAffected(); n == 0 {
			return fmt.Errorf("%w: lock %s is not live at epoch %d", ErrStaleEpoch, name, epoch)
		}
		return nil
	})
}

// lockHeldAt says whether the lock is live at that epoch.
func lockHeldAt(ctx context.Context, q querier, name string, epoch int64) (bool, error) {
	var n int
	err := q.QueryRowContext(ctx, `SELECT count(*) FROM locks WHERE name=? AND epoch=? AND expires_at > `+nowSQL, name, epoch).Scan(&n)
	if err != nil {
		return false, fmt.Errorf("ledger: %w", err)
	}
	return n > 0, nil
}

// reapLocks expires the locks whose holder alive judges Dead, outside any transaction for
// the judging (I3) and checking in the transaction that the lock is still the one judged.
func (s *Store) reapLocks(ctx context.Context, alive func(pid int, start string) Liveness) error {
	rows, err := s.db.QueryContext(ctx, `SELECT name,epoch,pid,pid_start FROM locks WHERE expires_at > `+nowSQL+` AND pid IS NOT NULL`)
	if err != nil {
		return fmt.Errorf("ledger: %w", err)
	}
	type lk struct {
		name  string
		epoch int64
		pid   int
		start string
	}
	var all []lk
	for rows.Next() {
		var l lk
		var start sql.NullString
		if err := rows.Scan(&l.name, &l.epoch, &l.pid, &start); err != nil {
			rows.Close()
			return err
		}
		l.start = start.String
		all = append(all, l)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	var dead []lk
	for _, l := range all {
		if alive(l.pid, l.start) == LivenessDead {
			dead = append(dead, l)
		}
	}
	if len(dead) == 0 {
		return nil
	}
	return s.write(ctx, func(tx *sql.Tx) error {
		for _, l := range dead {
			if _, err := tx.ExecContext(ctx, `UPDATE locks SET expires_at=`+expiredSQL+` WHERE name=? AND epoch=?`, l.name, l.epoch); err != nil {
				return fmt.Errorf("ledger: %w", err)
			}
		}
		return nil
	})
}
