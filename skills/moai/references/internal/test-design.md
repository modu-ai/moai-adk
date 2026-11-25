# Internal: test-design

When to read: before you write or change a test, as `moai-impl` writing the test first or as `moai-audit` judging whether a test proves anything.

1. Write the test first and run it red. Confirm it fails for the reason you expect before you change the code.
2. Test every enforced limit at its edge: the last value that passes (N) and the first that fails (N+1), and the empty case.
3. A guard needs a test in the direction it lets through and in the direction it refuses; a change that narrows what a check matches carries a literal test in the miss direction.
4. At audit tier 2 only, prove each new guard test with a mutation: commit first, then change the guarded code or delete the pinned sentence itself — not an introduction around it — run the test and expect FAILED, restore the bytes, and confirm the tree is clean before the next mutation. A mutation that survives means the test pins the wrong thing. Below tier 2 no mutation is asked for (`internal/self-check`, item 2).
5. A run that selected no test (`0 passed`) is not_run, never a pass. A test-name filter applies to every target of the call, so a filtered run names exactly one target.
6. Tests must not depend on the clock, the network, the real HOME or another worktree's build folder.
7. An expected value comes from the source of truth (the constant, the function or the file the product reads), never from a literal copied into the test: changing the source then changes what the test expects.
8. A test of product behaviour is written in the project's own test language; no second language is brought in for tests.
9. A test that must show code stays within a time bound (a scan that must stay linear) decides it by a count of work and not by a wall clock, because seconds change with the machine's load and the count does not: a counter that exists only in the test build adds every pass, every byte copied or lower-cased and every call into a shared routine, on the thread that runs the code so tests running side by side do not share it; the test sets a limit from the input length (the measured linear factor plus a small margin) and the counter stops the code at the limit, so a regression fails at once and does not wait; the test also fails when the counter counted less than the input length, so a counter that nothing feeds cannot pass; prove the whole with a mutant that makes the code quadratic and expect the test to fail on it. The count sees only the work that calls the counter, so a new copy or pass in code that does not call it stays invisible: add the call there too.
