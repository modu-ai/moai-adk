# Internal: boundary-checks

When to read: when a change crosses a boundary between two parts that were written or tested apart (a producer and its consumer, a file layout and the code that names its paths, a state table and the code that moves between states), as `moai-audit`, or as `moai-impl` writing the test for it.

Defects at a boundary pass each side's own tests and fail only when the two meet. Read both sides together; testing each side alone does not find them.

1. **List the boundaries** the change touches: data passed between components, paths and names one side writes and the other reads, states one side defines and the other sets.
2. **Read both sides in one pass.** Take the producer's actual output and the consumer's expectation side by side, not each against its own documentation.
3. **Trace the data** from where it is made to where it is used, through every transformation, and check each step:
   - shape: wrapped or unwrapped, list or object, which fields are required;
   - names: the same field or key spelled the same way on both sides (case style, prefixes);
   - every stage of an asynchronous or multi-step result, not only the final one.
4. **Check references against what exists.** Every path, route, command or identifier one side names must exist on the other side, with the same signature.
5. **Check state tables both ways.** Every transition the table allows has code that performs it, every state change in the code is in the table, and every intermediate state has a way out.
6. **Distrust casts and unchecked conversions.** A cast or an untyped parse lets a mismatch compile. Name each one and prove it with a test that uses real data from the other side.
7. **Test across the boundary.** The test drives both sides together and fails when they disagree. State the assumptions it relies on.

Adapted from revfactory/harness `qa-agent-guide.md` (Apache License 2.0), as carried in the MoAI ADK Go generation `rules/moai/quality/boundary-verification.md` with its `rules/moai/NOTICE.md` at revision `4fa2aa9db2`; the framework-specific case studies were generalised.
