---
title: /moai goal
weight: 25
draft: false
---

**条件声明型自主循环** 命令:用 `/moai goal "<条件>"` 声明完成条件后,会话会自主工作直到条件成立。每个回合结束时,`stop-goal` Stop 钩子评估条件是否满足,未满足就自动开始下一个回合。用 `/moai goal status [--all]` 查看进度,`/moai goal clear` 解除循环,`/moai goal resume` 从中断处继续。该命令没有独立的斜杠命令文件,而是通过 `moai` 技能路由和 `moai goal` CLI 进入的程序化命令面。
