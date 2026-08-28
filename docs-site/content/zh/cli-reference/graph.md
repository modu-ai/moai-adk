---
title: moai graph
weight: 16
draft: false
new: true
added_in: "v3.1.1"
---

{{< new-badge v3.1.1 >}}

把代码库里的关系收进一份产物、用来回答**反向提问**的工具。“改这个包会波及到哪里”、“这个 SPEC 和实际代码连着吗” —— 这类问题 grep 答不了，必须把关系聚在一起才能回答。

{{< callout type="info" >}}
**一句话总结**: `moai graph build` 把散落在 codemaps、@MX 标签、SPEC、报告里的关系收进 `.moai/project/graph/edges.jsonl` 一个文件，`moai graph query` 对这份文件做反向查询。
{{< /callout >}}

## 为什么需要它

MoAI-ADK 已经在多个层面持有关系信息 —— codemaps 的导入图、代码里的 `@MX:SPEC` 标签、SPEC 文档的依赖声明、报告里的里程碑记录。问题在于这些层面各自散落在不同的文件里。想问“改这段代码会影响哪个 SPEC”，就必须把导入方向（导入带 `@MX:SPEC` 标签文件的那些文件）和 SPEC 依赖方向**放在同一张图里**反向追。edges.jsonl 就是这同一张图。

## moai graph build

```bash
$ moai graph build
```

汇总导入边、`@MX:SPEC` 连接、SPEC 间依赖，写入 `.moai/project/graph/edges.jsonl`。它在同一个 git HEAD 上跑两次会产出相同内容 —— 是确定性的。查询永远读这份产物，所以**查询之前要先跑一次 build**。

在文档层之上，build 现在直接从代码中提取边——函数调用边(code-call)和导入边(code-import)，而既有的文档边一条都不变。导入目标会剥掉 go.mod 模块路径、归一化成仓库本地包，与 codemaps 的导入图指向同一个包域；每种语言的调用解析到什么水准(grade)对全部 16 种语言公开。两层对同一条关系说法不一致时，不丢弃任何一方：文档边带着 `disagrees_with` 标记留下，`--all-disagreements` 连被抑制的方向(代码发现了、文档沉默的本地依赖)也一并显示。

## moai graph query

一次调用只给**恰好一个**选择器。

| 选择器 | 提问 | 回答 |
|--------|------|-----|
| `--callers <节点>` | 谁直接依赖这个包/SPEC？ | 反向邻居 —— 导入它的包、依赖它的 SPEC、带 `@MX:SPEC` 标签的代码文件 |
| `--blast <节点>` | 从这里改起会波及多远？ | 沿反向边广度遍历 (BFS) 得到的影响半径。`@MX:SPEC` 边双向传播，能触及代码文件所实现的 SPEC |
| `--fanin [--limit N]` | 被用得最多的包是哪些？ | 按导入扇入排名 —— @MX:DEBT 扇入查询的代用品（还没有按标签种类分的边） |
| `--specs-no-code` | 哪些 SPEC 没有和代码相连？ | edges.jsonl 中 `@MX:SPEC` 边为 0 条的 SPEC 清单 |
| `--milestones-no-card` | 哪些里程碑没有卡片就过去了？ | 卡片交叉核对行未主张卡片、或主张的卡片不在存活积压队列里的里程碑 |

```bash
$ moai graph query --callers SPEC-FOO-001
$ moai graph query --blast internal/config
$ moai graph query --fanin --limit 20
$ moai graph query --specs-no-code
$ moai graph query --milestones-no-card
```

可以用 `--edges <路径>` 指向别的 edges.jsonl，或用根参数指定别的项目根。

查询作答前，如果机械层(@MX 索引 · edges.jsonl)已经过期，会先刷新再回答。只有内容哈希变了的文件才会重新解析，因此未提交的编辑也会反映在答案里；刷新耗时超过 gate.yaml 的 `update_budget_ms`(默认 2000ms)时只警告，答案照常给出。每条答案都会在 stderr 上打印计算它所用的树根与提交(或 dirty 指纹)，不会混淆答案属于哪棵树。

## moai graph check

```bash
$ moai graph check
codemaps  metric=described-source-diff value=0 threshold=40 verdict=fresh
mx-index  metric=inventory-content-diff value=0 threshold=1 verdict=fresh
edges     metric=source-fingerprint-mismatch value=0 threshold=0 verdict=fresh
```

按层各自的指标，测量图的三个层——codemaps · @MX 索引 · edges.jsonl——落后代码多远，并给出每层 `fresh` / `stale` / `absent` 判定。codemaps 看打了戳记的生成提交之后变化的被描述文件数(回退的改动计为 0)，@MX 索引看内容哈希变了的文件数，edges.jsonl 看源指纹是否不一致。

每个生成物都用 provenance 块声明自己描述的是哪棵树、哪个提交。没有这个块的产物判为 `absent`——判断不能冒充 fresh，absent 同样算失败：新 worktree 里这些未跟踪产物根本不存在，检查会如实说明而不是放行。退出码为 0(全部 fresh)· 1(stale 或 absent)· 2(系统错误)，提交前质量门的 graph-freshness 步骤和 CI 的 graph-freshness 作业直接消费这个值。阈值在 gate.yaml 的 `graph_freshness` 小节调整。

任何地方都不读 mtime。新检出会把所有 mtime 重置，基于 mtime 的指标会误判成刚重新生成——所以这里的指标只有内容哈希、git diff 和指纹。

stale 判定现在会附带归因信息。`codemaps` 层判为 stale 时，stderr 会先打印这次变更自身贡献的漂移文件数(`contribution`)和测量所依据的提交(`contribution_base`，通常是 `HEAD^1`)，接着列出最多 10 个引发漂移的路径，超出部分归纳为 `... and N more`。`--json` 中同样的信息以 `contribution` · `contribution_base` · `driving_paths` · `driving_paths_omitted` 字段暴露。一行 stderr 就能分清这条 lane 是仅仅继承了这个 red(贡献 0)还是自己造成的(贡献 > 0)。

## moai graph stamp codemaps

```bash
$ moai graph stamp codemaps
OK: stamped .moai/project/codemaps/provenance.json
provenance: tree=/path/to/project commit=1a2b3c4d5e6
```

重新生成 codemaps 后，作为最后一步执行。文档内容由 `/moai codemaps` 打磨，而这份内容**描述的是哪个树状态**由本命令写入 `provenance.json`——`moai graph check` 判定 codemaps 层的依据就是这份记录。

### 指定一个能活过合并的提交（`--commit`）

```bash
$ moai graph stamp codemaps --commit "$(git merge-base HEAD origin/main)"
OK: stamped .moai/project/codemaps/provenance.json
provenance: tree=/path/to/project commit=1a2b3c4d5e6
```

不带旗标执行时，戳记记录的是当前检出的 HEAD。在功能分支上这是个陷阱：本仓库用**挤压合并**（squash merge）合入拉取请求，分支上的提交——HEAD 也在内——永远不会进入 main 的历史。指向分支本地 HEAD 的戳记在挤压合并落地的那一刻就成了孤儿，之后打开的每个拉取请求都会继承 graph-freshness 的红灯（`not comparable`，exit 2）。这个失败真实发生过一次，并以 `0d15864ae90b` 事件留档追查。

`--commit <rev>` 接受任何 `git rev-parse` 表达式（完整 sha、短哈希、引用名都可以），解析成完整 sha 后原样记录。上面的 merge-base 配方就是安全写法：`git merge-base HEAD origin/main` 既是 main 的祖先（能活过挤压合并），内容又与分支点上的 described 源一致（不会把别的拉取请求合入的改动算成你的漂移）。千万不要对着分支本地 HEAD 重新盖戳。另外，described 源存在未提交改动时使用 `--commit` 会被直接拒绝——指定提交和内容指纹是两种不同的诚实声明，而 schema 里锚的位置只有一个。

这条纪律由 CI 机械地兜底：graph-freshness 工作流在给出任何新鲜度判定之前，先验证被跟踪戳记的提交是拉取请求目标分支的祖先。注定成为孤儿的戳记会在原地点名失败，而不是等合并之后才以一个无名的 exit 2 冒出来。

## 两个选择器的注意事项

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--specs-no-code`**: “未连接”不等于“未实现”。大多数 SPEC 只产出文档、规则、挽具，没有代码也算完成。请把这份清单当作覆盖度地图来读，而不是缺陷清单。
{{< /callout >}}

{{< callout type="warning" >}}
{{< icon warning warn >}} **`--milestones-no-card`**: 积压队列在卡片结束（`done`）时会删掉对应行。所以“不在队列里”**同时**混着已结束的卡片和从未签发的卡片。请用 `git log --oneline --grep 'merge: tNN'` 逐项判别 —— 有提交即通过，没有则是新卡片候选。grep 到 0 条也不代表“没干过活”：卡片可能换了新 id 重新签发，切断新卡片之前先核对谱系。
{{< /callout >}}

## 相关文档

- [看板模式](/zh/advanced/kanban-mode) — 里程碑-卡片交叉核对所守护的卡片流
- [`/moai mx`](/zh/utility-commands/moai-mx) — @MX 标签与 `@MX:SPEC` 连接的源头
- [Navigator](/zh/core-concepts/navigator) — 把设计决策、SPEC、符号绑成一张图的另一处图 (nav-graph.json)
