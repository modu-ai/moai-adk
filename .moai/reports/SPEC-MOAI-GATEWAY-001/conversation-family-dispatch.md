# Retained conversation family implementation dispatch

Parent attribution: developer-provided source_session_id 01a08e7b-6aa0-7361-ab7e-ea8da1f02228. WT current CLI reports environment-fallback. Work only in moai-proxy-unified, WT-unified-gateway, HEAD 81c1d58f9 plus existing uncommitted implementation. No remote operations, commits, SPEC edits, primary-checkout edits, or unrelated cleanup.

## Approved source and ownership

Implement the approved core design.md RPA-1 retained native profile and preauthorized lifetime table (currently lines 498–540), paired with acceptance and PICKER requirements. Read these sources and receipt APIs before writing. Own only a new internal/gateway/conversation package, focused tests and conversation-family-verification.md. Other actors own opaque codec, gateway policy adapters, AUTH Windows and audits. Do not revert their work or silently widen ownership.

This component provides concrete retained family storage, owned conversation index, family process lease and pre-launch UUID resolution. It must not activate the root launcher or claim native/provider integration. Prefer existing private filesystem/locking/atomic primitives; do not create an independent database or duplicate security machinery without explaining the required seam.

## Required behavior

- New conversation: trusted caller-supplied actual native CWD/project identity; generate UUID before child; create private family native config plus initial receipt, then expose a launch descriptor. No credential or original transcript copying.
- Exact resume: UUID must belong to the selected project and current user. Validate family, receipt and completed native transcript mapping before returning exact resume arguments. Missing state is an error, never an empty new family.
- Continue: resolve the owned index's last successful completion to exact UUID; no ambiguous native --continue. Selection: enumerate only validated owned entries, return descriptors for a caller-selected UUID; UI remains a later CLI integration.
- Fork: snapshot parent's validated hash receipt into child manifest, preserve same native family config, return --resume parent --fork-session --session-id child. Pin inherited candidates independently of parent future updates/deletion. Partial publication must not produce a resumable child.
- One native lead per family through a process-owned lease. Same family busy must fail explicitly; independent families run concurrently. Crash releases lease, but retained native config/receipt/index survives ordinary Close.
- Do not register completion until trusted native transcript inside the private root has exact UUID/project identity and complete record evidence. Reject symlinks, wrong owner/mode/identity, duplicate or ambiguous mappings and incomplete transcripts. Bound file sizes/counts. Store only family/UUID, normalized project identity, native transcript relative identity and completion sequence; never public conversation, opaque or tokens in the index.
- Preserve native actual CWD. Do not reuse launchProjectRoot blindly: that helper is the goal-state project resolver and may canonicalize the worktree differently. Transcript evidence scripts discover projects/**/<UUID>.jsonl under the private config. Determine and validate actual transcript schema from synthetic preflight captures; do not guess a path encoding or copy user transcripts.
- Secure namespace selection follows approved design exactly: existing CLAUDE_SECURESTORAGE_CONFIG_DIR including empty preserved; otherwise original nonempty CLAUDE_CONFIG_DIR; otherwise explicit empty secure override. Original explicitly empty CLAUDE_CONFIG_DIR is unsupported preflight. Retained CLAUDE_CONFIG_DIR is separate. No keychain/token reading by this component.
- Explicit discard only; reject active family, retain live fork siblings, never broad recursive deletion on uncertain mapping. Do not expose an unrequested CLI command here.

## Evidence and constraints

Native fork and exact resume positive evidence is in native-fork-runtime-observation.md and picker-resume-runtime-observation.md. Secure namespace live positive is picker-settings-source-observation.md. These are installed-client observations, not product completion.

Receipt API is implemented under internal/gateway/receipt with bounded hash-only candidates and POSIX Store; iter3 fixed cancellation/lock/ambiguity probes passed. Windows receipt currently rejects unsupported platform. Windows AUTH primitives have separate approved API semantics; do not silently claim Windows conversation support from cross-compilation. Report any shared primitive extraction needed as a concrete integration seam, without writing another actor's files.

TDD on meaningful lifecycle cases: new/resume, last-completed ordering and ties, fork snapshot independence, concurrent family isolation/busy, process crash release, missing/corrupt state, wrong project/UUID, path traversal and symlinks, incomplete transcript, secure namespace matrix, cancellation/partial publication. Run scoped race/vet and cross-compile where relevant. Report command plus original output, measured source hashes, baseline, gaps and residual risk. Return ownership when this component is complete; integration and independent audit follow separately.
