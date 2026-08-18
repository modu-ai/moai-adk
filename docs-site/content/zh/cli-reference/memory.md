---
title: moai memory
weight: 18
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

教训存储库（Lessons Protocol memory store）的**健康诊断与归档**工具。MoAI 的教训由 `MEMORY.md` 索引和一组主题文件构成，而会话开始时读的是索引、不是整个目录。于是只有文件、没有索引行的教训就成了**存下了却永远不会被回想起**的状态。

{{< callout type="info" >}}
**一句话总结**: `moai memory doctor` 诊断索引↔文件的不一致（孤儿文件·空链接）和主题文件数量上限，`moai memory archive` 把点名的文件折进归档并从索引摘下。
{{< /callout >}}

## moai memory doctor

```bash
$ moai memory doctor            # 供人阅读的报告
$ moai memory doctor --json     # 结构化输出
$ moai memory doctor --dir <路径>   # 诊断别的存储库
$ moai memory doctor --cap 80   # 换一个上限来检查
```

诊断项:

| 项 | 含义 |
|------|-----|
| 孤儿主题文件 | 没有索引行的主题文件 —— 不会被回想的教训 |
| 空索引链接 | 索引行指向的文件不存在 —— 留下读取失败的行 |
| 主题文件数 | 相对项目上限（默认 50）的当前数量 |

`--dir` 和 `--cap` 只是改变检查对象与基准的选项，不会修改文件。doctor 只做诊断。

## moai memory archive

```bash
$ moai memory archive feedback-old-lesson.md
```

把点名的主题文件移到 `memory/_archive/` 并摘下索引行。**不是删除** —— 归档保留审计痕迹。什么算过时属于判断问题，所以由操作者按文件名亲自点名，没有自动筛选。它的主要用途是超过上限（默认 50 个）时把超出部分折进归档，保持索引轻量。

## 相关文档

- [上下文与记忆](/zh/claude-code/context-memory/memory) — Claude Code 记忆的工作原理
- [自我进化](/zh/advanced/self-evolving) — 教训上升为规则晋升提议的流程
- [决策记忆](/zh/advanced/decision-memory) — 路由决策沉淀的另一处记忆
