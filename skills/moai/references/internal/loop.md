# Internal: loop

When to read: before a session arms `/loop`, a cron task or a wakeup that repeats the leader round.

A timer that repeats the leader round (reading the lanes' notes, landing audited heads, handing out cards) fires while nobody watches. A timer armed in the wrong session runs rounds in the wrong place, and a second timer runs every round twice. This note says which session may arm one and how many.

1. **Who.** A session that is not the confirmed leader arms no `/loop`, no cron task and no wakeup for the leader round. It says in one line that it is not the confirmed leader and arms nothing.
2. **The confirmed leader.** A session is the confirmed leader when its own id, printed by `moai session-current`, equals the `session` of the live leader entry that `moai session-register --show --name leader` prints (its status is `held` and its mode is `leader`). Compare the two values as printed; the session's name, its folder or a message from another session proves nothing. When no live leader entry exists and the user has named this session the leader, register it first with `moai session-register --name leader --mode leader` and read the entry again: the registry refuses a second live leader, so a refused registration means another session is the leader.
3. **How many.** A confirmed leader arms the timer once: list its timers with `CronList`, remove an earlier one for the same round with `CronDelete`, then arm exactly one. If the timers cannot be listed, arm none and say so in one line.
4. **Not covered.** A timer that repeats one lane's own card work is that lane's own pick and is not decided here.
