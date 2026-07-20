---
title: 提示缓存 —— 概念与在 Claude Code 中的行为
weight: 30
draft: false
---

提示缓存是一项 API 功能：**当请求的前段（前缀）与上一次请求相同时，不再重新处理该部分，而是复用它**。从缓存读出的令牌按基础输入单价的 **0.1 倍**计费，因此重复出现的上下文（系统提示、项目指令、对话历史）越大，节省效果越明显。如果说 MoAI-ADK 代币经济学的"上下文瘦身"这条轴是减少常驻加载的上下文，那么提示缓存就是把剩下的上下文廉价地复用。

{{< callout type="info" >}}
**通俗比喻** —— 模型在每一回合都会从头重读整段对话。缓存就是一枚书签，说一声"前段还是刚才读过的那样"便跳过。前段哪怕改动一个字，书签就失效，要从那一点起重读。
{{< /callout >}}

## 核心概念（各 API 通用）

- **前缀匹配**：要命中缓存，直到断点为止的内容（含工具定义·系统提示·消息历史）必须**100% 相同**。哪怕只差一个空格，从那一点之后就全部重算。
- **价格倍数**（相对基础输入单价）：缓存写入 5 分钟 TTL **1.25 倍** · 1 小时 TTL **2 倍** · 缓存读取 **0.1 倍**。
- **TTL（存活时长）**：默认 5 分钟，可选 1 小时。在 TTL 内每被复用一次，存活时长就免费延长。
- **各模型的最小缓存令牌数**：比这更短的前缀不会被缓存（不报错，按普通方式处理）。例如：Fable 5 = 512、Opus 4.8·Sonnet 5 = 1,024、Opus 4.7 = 2,048、Haiku 4.5 = 4,096 令牌。

## 如果你是 Claude Code 用户 —— 缓存是自动的

Claude Code **自动管理**提示缓存。你既不需要、也无法亲手设置 `cache_control`。以官方文档为准，其行为如下：

- **TTL 自动选择**：在订阅套餐（Pro/Max/Team/Enterprise）下会自动请求 1 小时 TTL（已含在套餐费用中，无额外成本）。经由 API 密钥·云服务商时默认 5 分钟，可用 `ENABLE_PROMPT_CACHING_1H=1` 选择加入 1 小时。
- **环境变量控制**：`FORCE_PROMPT_CACHING_5M=1`（强制 5 分钟）、`DISABLE_PROMPT_CACHING=1`（全面禁用 —— 除调试外不推荐）、按模型的 `DISABLE_PROMPT_CACHING_OPUS` 等。
- **请求构成优化**：Claude Code 会把不常变的内容（系统提示 → 项目上下文 → 对话）排在前面，以提高前缀命中率。

### 会使缓存失效的行为（该回合变慢且变贵）

- 切换模型（`/model`）· 变更 effort（`/effort`）—— 缓存按模型·effort 分开
- 连接/断开 MCP 服务器（当工具定义已加载进前缀时）
- 添加/移除工具的整体拒绝（deny）规则
- `/compact`（对话历史被替换为摘要）· Claude Code 升级后的第一回合

### 会保持缓存的行为

- 编辑仓库文件（只在读取时才加入对话）· 调用技能/命令 · 切换权限模式 · `/rewind`（回到已缓存的前缀）· 会话途中修改 CLAUDE.md（缓存保留，但**改动也不会生效** —— 在下一次 `/clear`·重启时反映）

### 查看命中率

- 响应中的 `cache_creation_input_tokens`（写入）/ `cache_read_input_tokens`（读取）比例就是指标 —— 读取越高说明运行得越好。
- 通过 MoAI statusline 的 cache_hit 段可在会话中实时查看。
- 如果每回合的写入都持续偏高，说明上面"会使缓存失效的行为"中有某项正在改变前缀。

## 直接调用 API 时（参考）

以下用法示例仅适用于**直接调用 Anthropic API 的开发者**。Claude Code 用户不适用。

```python
# 直接调用 API：在稳定的系统提示上放置缓存断点
response = client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system=[{
        "type": "text",
        "text": "稳定的系统提示...",
        "cache_control": {"type": "ephemeral", "ttl": "1h"}
    }],
    messages=[{"role": "user", "content": user_query}]
)
```

原则只有一条：**把断点放在每次请求都会变的数据（问题、时间戳）之前的最后一个稳定块上**。盈亏平衡点是 2 个请求 —— 第一个请求的写入溢价，会由 TTL 内第二个请求的 0.1 倍读取收回。

## MoAI cache.yaml 的适用范围

`.moai/config/sections/cache.yaml`（`enabled`、`session_ttl`）只适用于 **MoAI 经自有 SDK 包装路径直接调用 Anthropic API 时的 cache_control 注入**。**它与 Claude Code 会话的缓存无关** —— Claude Code 的缓存如上节所述由运行时自动管理，MoAI 无法介入。

> **GLM 后端**：z.ai（GLM）使用基于内容相似度的**隐式缓存**，因此 MoAI 不在 GLM 路径注入 `cache_control`。

## 小结

- **Claude Code 用户**：无需任何设置。只要在工作之间的自然边界上才切换模型/effort 与执行 `/compact`，命中率就能保持。
- **直接调用 API 者**：在稳定块上放断点，只有在 2 个以上请求时才用 1 小时 TTL。
- **监控**：statusline cache_hit 段 + `cache_read/creation` 令牌比例。

**来源（官方文档）：**

- [How Claude Code uses prompt caching](https://code.claude.com/docs/en/prompt-caching) —— 自动管理、TTL 自动选择、失效/保持行为、环境变量
- [Prompt caching (API)](https://platform.claude.com/docs/en/build-with-claude/prompt-caching) —— cache_control、价格倍数、各模型最小令牌数
- [Manage costs effectively](https://code.claude.com/docs/en/costs) —— Claude Code 的自动成本优化
