# Internal: test-design

When to read: before you write or change a test, as `moai-impl` writing the test first or as `moai-audit` judging whether a test proves anything.

1. Write the test first and run it red. Confirm it fails for the reason you expect before you change the code.
2. Test every enforced limit at its edge: the last value that passes (N) and the first that fails (N+1), and the empty case.
3. A guard needs a test in the direction it lets through and in the direction it refuses; a change that narrows what a check matches carries a literal test in the miss direction.
4. Prove each test with a mutation. Commit first, then change the guarded code or delete the pinned sentence itself — not an introduction around it — run the test and expect FAILED, restore the bytes, and confirm the tree is clean before the next mutation. A mutation that survives means the test pins the wrong thing.
5. A run that selected no test (`0 passed`) is not_run, never a pass. A test-name filter applies to every target of the call, so a filtered run names exactly one target.
6. Tests must not depend on the clock, the network, the real HOME or another worktree's build folder.
