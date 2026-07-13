---
title: /moai goal
weight: 25
draft: false
---

A **condition-declared autonomous loop** command: declare a completion condition with `/moai goal "<condition>"` and the session keeps working on its own until the condition holds. After each turn the `stop-goal` Stop hook evaluates whether the condition is met and starts another turn if it isn't. Use `/moai goal status [--all]` to inspect progress, `/moai goal clear` to release the loop, and `/moai goal resume` to continue from where it left off. There is no dedicated slash-command file — entry is via the `moai` skill router and the `moai goal` CLI, making this a programmatic command surface.
