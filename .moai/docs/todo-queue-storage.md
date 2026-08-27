# Backlog queue storage and the downgrade route

The backlog queue `moai todo` operates lives in one SQLite database at
`.moai/state/todo/backlog.db`. This page explains what each artifact in that
directory is, how an existing project moves onto the database, and how to get
back to plain JSON if you need to run an older release.

Nothing here is required for normal use. Read it when you are downgrading,
recovering from an interrupted upgrade, or wondering what a file in that
directory is for.

## The artifacts

| Artifact | What it is | Safe to delete? |
|---|---|---|
| `backlog.db` | The live queue. Cards, findings, and the id high-water mark. | **No.** This is the queue. |
| `backlog.db-wal` | SQLite's write-ahead log. Holds committed data not yet folded into the main file. | **No** while anything is running. Deleting it mid-write loses recent cards. |
| `backlog.db-shm` | SQLite's shared-memory index for the WAL. Rebuilt automatically. | Only when nothing is using the queue. |
| `backlog.lock` | The advisory lock every writer takes. Lets several sessions share one queue without losing updates. | Yes when nothing is running; it is recreated on demand. |
| `backlog.json` | Present only if you exported one (see below), or if the queue has not been moved onto the database yet. | Yes, once you no longer need it — but see the downgrade route first. |
| `backlog.json.migrated` | Your original JSON queue, preserved byte-for-byte at the moment it moved onto the database. | Yes, once you are confident in the upgrade. Keeping it costs a few hundred kilobytes and is the simplest rollback there is. |

## Moving an existing queue onto the database

It happens by itself, once, the first time a `moai todo` command runs against a
project that still has a `backlog.json`. You do not run anything.

The order matters and is worth knowing, because it is what makes the move safe
to interrupt:

1. The queue lock is taken, so concurrent sessions wait rather than race.
2. The JSON is read in full. A file that will not parse **stops here** — the
   move is abandoned and your JSON stays exactly as it was. Nothing is
   repaired, rewritten, or guessed at.
3. The database is written in one transaction.
4. It is read back and compared to the JSON **field by field** — every card,
   every finding, in order, plus the id high-water mark.
5. Only if that comparison passes is the JSON renamed to
   `backlog.json.migrated`.

Because the comparison happens before step 5, a move that would have lost
anything leaves the JSON in place and authoritative. Your queue keeps working
on the old file and the command reports what went wrong.

If the process is killed between steps 3 and 5, you are left with both files.
The database is authoritative from that point on, and the next `moai todo`
command finishes the rename. Nothing is lost either way.

## Downgrading to a release that predates the database

An older `moai` reads only `backlog.json` and ignores the database entirely, so
the whole job is producing a current JSON file for it to read.

```sh
moai todo export-json     # writes .moai/state/todo/backlog.json from the live queue
```

Then install the older release. It will pick the exported file up as its queue.

Three things worth knowing:

- **Export last.** The file is a copy taken at the moment you run the command,
  not a live mirror. Cards added afterwards are in the database only, so export
  immediately before you swap binaries.
- **The export is left alone.** Later `moai todo` commands do not rename or
  remove it, even though the database is still the queue they read and write.
- **It does not undo the move.** The database stays authoritative for this
  release. `export-json` produces a file for a different binary to read; it
  does not switch this one back.

There is deliberately no setting that selects the storage engine. Two live
engines would mean two places a card could be, and the whole point of one
store is that there is only ever one answer to "where are my cards?".

## Recovering

**"I exported, downgraded, and now want to come back."** Just install the newer
release. It finds the `backlog.json` the older one has been using and moves it
onto the database again, with the same verification described above.

**"Something is wrong with the database."** A queue that cannot be read is
reported as an error — never as an empty queue, which would look like your
cards were gone. The database is never deleted or rewritten in response.
Your options, in order of preference:

1. If `backlog.json.migrated` is still there, it is your queue as of the
   upgrade. Rename it to `backlog.json`, move `backlog.db*` aside, and the next
   command moves it back onto a fresh database.
2. If it is not, `backlog.db` is still a standard SQLite file and standard
   SQLite tools can read it.

## The directory name

This directory is `.moai/state/todo/`. Earlier releases called it
`.moai/state/kanban/`, which named no command anyone could type — the queue has
always belonged to `moai todo`.

The rename happens automatically, once, on the first `moai todo` command, and
moves the whole directory: the queue and the per-session records that live
beside it travel together.

Two cases are worth knowing:

- **If both directories exist**, the new one wins and the old one is left
  exactly where it is, untouched. It is left visible on purpose. If some script
  of yours is still writing to the old path, you want to be able to see that
  rather than have it quietly swallowed.
- **If the directory cannot be moved** — a permission you did not expect, or a
  mount boundary between the two paths — the queue is served from the old
  location instead and nothing fails. It is retried on the next command.

Read-only surfaces (the web console, the status line) never trigger either the
directory move or the storage move. They read whichever layout they find.
