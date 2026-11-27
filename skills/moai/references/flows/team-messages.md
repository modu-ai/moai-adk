# Team messages

How the leader window and the agents (lanes) talk during one card. A session started with `moai cc -l` is named `leader-<project name>` and one started with `moai cc -a` is named `agent-<english name>`, a name no other session on the machine holds (`moai session-register --names` prints the project name); a session started plain takes the same names with `/rename` (`workflows/work.md`, `workflows/leader.md`). Every step of the flow goes by a cross-session message, and the ledger is the record.

## The message table

Every message is one line whose head is `<KIND> <card> <state> head=<10 characters>`, followed by what the receiver needs in the user's conversation language. Commands, paths, card ids and protocol lines stay verbatim.

| Message | From | To | Head |
| --- | --- | --- | --- |
| Dispatch | leader | agent | `DISPATCH <card> ready head=<integration tip>` and the scope; the instruction lives in the message, never after `/moai agent` |
| Question | agent | leader | `QUESTION <card> blocked head=<its head>`; the same text is a `LANE-BLOCKED kind=decision` note on the card |
| Answer | leader | agent | `ANSWER <card> <state> head=<its head>`; the decision is a note on the card |
| Audit request | leader | auditor agent | `AUDIT-REQUEST <card> ready head=<head to audit>`, the caps and the receipt to cite |
| Verdict | auditor agent | leader | `VERDICT <card> <pass or fail> head=<audited head>`; the verdict is recorded with `moai card verdict` |
| Repair request | leader | author agent | `REPAIR <card> fail head=<audited head>`; the findings are a note on the card |
| Landing notice | leader | agents concerned | `LANDED <card> landed head=<landed head>` after `moai card land` answered |
| Completion | agent | leader | `LANE-DONE <card> <state> head=<its head> (<name>)` where the state is READY, REPAIRED, AUDIT PASS, AUDIT FAIL or BLOCKED; once, after the note |
| Next card | leader | agent | a `DISPATCH`, sent in the same round once the ledger shows the previous card done or released |

## A message is a notice

A message is a notice and the ledger is the record. The receiver reads the ledger before it acts and acts on what the ledger shows; a message the ledger does not back starts nothing. One message per event, no retry loop.

## Agent to agent

An agent may send a message to another agent to check a fact, for example which head an audit read. A verdict, a decision and a repair request go to the ledger as well, because the ledger is what the leader and the auditor read.

## Addressing

Address the registered leader by the session id that `moai session-register --show --name leader` answers, matched to a row of `ListAgents`. Before you send a message to any session other than the registered leader, run `moai session-register --lane-check <listed name>` and send only when it answers `lane` (the name's live lock is held or its registry entry has a running lease); on `not_lane` send nothing and write one ledger note that names the row and the answer. When two rows of `ListAgents` carry the same name, send to neither and write one ledger note. Address a checked session by its listed name. Never send to a row picked by the bare name `leader`: another project's session can carry it.

## Channel order

Use the host's cross-session message first. When that channel is missing, refused or held, use the lane-terminal manager's own messaging (its adapter reference under `references/adapters/` names the send command, and the check command for the leader). When neither exists, write the ledger note alone. `/clear` and screen reading stay with the lane-terminal manager.
