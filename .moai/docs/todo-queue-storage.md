# Backlog queue storage and the downgrade route

The backlog queue `moai todo` operates lives in one SQLite database at
`~/.moai/db/<project-key>/todo/backlog.db`. This page explains what each
artifact in that home-scoped directory is, how an existing project is copied
into the database, and how to get back to plain JSON if you need to run an
older release.

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
| legacy project-local queue | The former `.moai/state/todo/` or `.moai/state/kanban/` source. It remains untouched after a verified import. | **No** until the home database has been backed up and verified in normal use. |

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
5. Only after that comparison passes does the home database become the queue
   returned by future path resolution. The project-local source is retained as
   a rollback snapshot; the migration does not delete or rename it.

Because the comparison happens before step 5, an import that would have lost
anything leaves the legacy queue in place and authoritative. Your queue keeps
working on the old store and the command reports what went wrong.

If the process is killed between steps 3 and 5, you are left with both stores.
The next `moai todo` command repeats the logical copy and readback; it does not
delete the source. Nothing is lost either way.

## Downgrading to a release that predates the database

An older `moai` reads only `backlog.json` and ignores the database entirely, so
the whole job is producing a current JSON file for it to read.

```sh
moai todo export-json     # writes backlog.json beside the home-scoped live database
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

## Factory assignment provenance

A factory run starts before it necessarily knows which SPEC a card will use.
For that reason, the run manifest records run metadata only. The append-only
`card.assigned` event written by `moai todo next <n> --spec <SPEC-ID>` is the
source of truth for the assigned SPEC snapshot: canonical `spec.md` path,
SHA256, Git commit, and capture timestamp. Missing SPEC files and non-Git
projects remain usable; the unavailable hash or commit is recorded as empty.

## Recovering

**"I exported, downgraded, and now want to come back."** Just install the newer
release. It finds the `backlog.json` the older one has been using and moves it
onto the database again, with the same verification described above.

**"Something is wrong with the database."** A queue that cannot be read is
reported as an error — never as an empty queue, which would look like your
cards were gone. The database is never deleted or rewritten in response.
Your options, in order of preference:

1. Stop every process using the queue and back up `backlog.db*` together.
2. The retained project-local queue is your rollback source as of the import.
   Move the home database backup aside, then let the next `moai todo` command
   import that source again.
3. `backlog.db` is a standard SQLite file, so standard SQLite integrity and
   recovery tools can inspect the backup without changing the live copy.

## The home directory

Each project has one stable key below `~/.moai/db/`. Linked worktrees derive
the key from the primary checkout, so every session reads and writes the same
queue. `project.json` beside `todo/` records the canonical project root used to
derive that key.

Read-only surfaces (the web console and status line) never trigger migration.
They read the home database when it exists and otherwise read the legacy
project-local queue. The first adopting `moai todo` command performs the
verified copy.
