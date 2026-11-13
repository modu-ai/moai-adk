# Internal: decide

When to read: before you ask anyone a question, or when a step needs a choice the task did not make.

The decision order is defined once, in the `moai` skill's `SKILL.md`; this note only says how to apply it.

1. Sort the question in the order `SKILL.md` gives: a fact that the code, a command or a document answers is not a decision; a choice that cannot be undone, touches one of the five HOLD categories (`money`, `legal`, `external_contract`, `secret`, `real_data`) or changes a `guarded` setting goes to the user and waits; a reversible choice outside those is decided within the autonomy level and recorded.
2. Answer a fact with its evidence — the command and its output, or the file and line — instead of asking.
3. Record every decision you make with the fields `SKILL.md` names: what was decided, the options, the default, whether it is reversible, its category, the deadline and who decided. A decision without a record does not count.
4. A lane does not decide for the user. It asks the leader; the leader asks the user only for what `SKILL.md` reserves to the user.
5. Never decide a HOLD, an independent audit or the CI gate.
