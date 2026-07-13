---
title: 提示词缓存 — 盈亏平衡分析与实现指南
weight: 30
draft: false
---

提示词缓存通过在多个请求中复用相同的提示词前缀，以 90% 折扣的成本降低
推理开销（基础成本的 0.1 倍）。如果说 MoAI-ADK 托克诺米克斯的"上下文
瘦身"轴是减少常驻加载上下文，那么提示词缓存就是把剩下的上下文便宜地
复用。本指南说明盈亏平衡规则、缓存机制，以及在 MoAI 项目中何时启用
缓存。

## 盈亏平衡规则

**1 小时缓存只在会话中发生 2 个及以上连续 API 请求时才启用。**

单请求会话（一次性查询、单轮命令）会产生 2 倍的写入溢价且没有缓存复用
收益，导致每请求成本反而高于不使用缓存的基线。相反，如果 1 小时内有多轮
交互或重复的 SPEC 分析，缓存会在第二个请求处抵消成本，并在后续请求中
节省 67% 以上。

### 成本比较

以 Claude Opus 4.5 为基准：

| 场景 | 无缓存 | 含 1 小时缓存 | 节省 |
|---------|----------|-----------------|-------|
| 1 个请求, 10K Token | $0.05 | $0.0625 | -25%（溢价） |
| 2 个请求, 10K + 10K | $0.10 | $0.0625 + $0.005 = $0.0675 | 节省 32% |
| 3 个请求, 10K + 10K + 10K | $0.15 | $0.0625 + 2×$0.005 = $0.0725 | 节省 52% |
| 5 个请求, 10K + 10K + 10K + 10K + 10K | $0.25 | $0.0625 + 4×$0.005 = $0.0825 | 节省 67% |

盈亏平衡点是 **2 个请求**：第一个请求的 2 倍写入溢价，可通过 1 小时 TTL 窗口内后续请求的 90% 折扣缓存读取收回。

## 缓存控制的工作原理

在提示词前缀上启用缓存控制后，缓存生命周期遵循以下模式：

1. **第一个请求（缓存写入）**：API 响应完成后，前缀被写入缓存。成本：`前缀_Token × 2.0（1 小时缓存）或 1.25（5 分钟缓存）`。
2. **后续请求（缓存读取）**：前缀相同且在 TTL 内时，从缓存中检索。成本：`前缀_Token × 0.1`。
3. **自动回溯**：系统会在最近 20 条消息中检查是否有缓存的前缀匹配。若找到则应用读取费率。

### 缓存放置最佳实践

将缓存控制断点放置在**每请求数据之前的最后一个稳定块**上：

```python
# 正确: 稳定的系统提示词（可缓存）
response = client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system=[
        {
            "type": "text",
            "text": "你是一名代码审查员...",
            "cache_control": {"type": "ephemeral", "ttl": "1h"}
        }
    ],
    # 下面是可变的每请求数据（不缓存）
    messages=[{"role": "user", "content": user_query}]
)

# 错误: 缓存断点在可变数据上
# 当前时间: {timestamp}
cache_control={"type": "ephemeral"}
# ^ 时间戳每个请求都变化，因此永远不会匹配
```

## 配置：session_ttl 与 spec_ttl

MoAI 缓存在 `.moai/config/sections/cache.yaml` 中配置：

```yaml
cache:
  enabled: false  # 要启用缓存请设置为 true
  session_ttl: "1h"  # 会话级缓存 TTL
  spec_ttl: "5m"     # SPEC 正文缓存 TTL
  min_cacheable_tokens: 2048  # 缓存写入最小 Token 数
```

### 通过 session_ttl: "off" 选择退出

要在特定会话中禁用缓存（例如一次性请求占主导时）：

```yaml
cache:
  enabled: true
  session_ttl: "off"  # 在此会话中禁用缓存
  spec_ttl: "5m"      # 只对 SPEC 正文使用缓存
```

当 `session_ttl: "off"` 时：
- 跳过会话级缓存写入
- 若配置了 `spec_ttl`，SPEC 正文缓存仍然生效
- 适用于以单请求为主的中断型工作流

## 监控缓存性能

使用 `moai doctor` 查看缓存命中率并决定是否启用缓存：

```bash
moai doctor --cache-metrics
```

输出示例：

```
缓存性能（过去 7 天）:
  缓存命中率: 67%
  总缓存读取: 450K Token
  总缓存写入: 150K Token
  节省: $2.15（相比无缓存节省 68% 成本）

单轮请求比率: 12%  ⚠️ 警告: 12% 的请求是单轮
                            （这些请求没有缓存收益）。
```

MoAI statusline 中也会显示缓存命中率 (cache_hit) 段，可在会话中即时确认
上下文瘦身与缓存配置的效果。

### 指标解读

- **命中率 > 60%**：缓存有效。保持启用。
- **命中率 30-60%**：中等收益。若会话以多轮为主，考虑启用。
- **单轮比率 > 30%**：缓存收益有限。确认"2 个及以上请求"的假设是否成立。
- **最小 Token 阈值警告**：配置 `min_cacheable_tokens` 以避免缓存小提示词（开销 > 节省）。

## 何时发生缓存未命中

缓存命中需要：

- ✓ 到断点为止的提示词前缀完全相同
- ✓ 在 TTL 窗口内（1 小时或 5 分钟）
- ✓ 相同的工作区/组织上下文
- ✓ 断点之前的所有块没有变化（工具、系统、顶层参数）

缓存未命中（常见原因）：

- ✗ 工具定义变化（工具参数不同）
- ✗ 网络搜索开关打开/关闭
- ✗ 前缀中添加或移除图片
- ✗ 扩展思考设置变化
- ✗ 断点之前的内容不同（包括空格）

## 最小 Token 阈值

缓存写入只在前缀超过各语言模型的最小值时才发出：

- **Claude Opus 4.5, 4.7, 4.8, Haiku 4.5**：最少 2,048 Token
- **Claude Opus 4.1, Sonnet 模型, 其他 Haiku 版本**：最少 1,024 Token

低于最小值的请求会在不缓存的情况下处理（无错误 —— 自动回退）。

## 预热（高级）

要消除第一次用户交互的缓存未命中延迟：

```python
# 缓存预热（用户到达之前）
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=0,  # 不计费输出 Token
    system="很长的系统提示词（5000 Token）...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": "warmup"}]
)
# 成本: system_tokens × $2.0/MTok（缓存写入）

# 之后: 用户请求命中已预热的缓存
client.messages.create(
    model="claude-opus-4-8",
    max_tokens=1024,
    system="很长的系统提示词（相同）...",
    cache_control={"type": "ephemeral", "ttl": "1h"},
    messages=[{"role": "user", "content": user_input}]
)
# 成本: system_tokens × $0.1/MTok（从预热读取缓存）
```

## 总结

- **启用缓存**：1 小时内有 2 个及以上连续 API 请求的会话
- **禁用缓存**：一次性查询或中断型工作流
- **监控**：用 `moai doctor --cache-metrics` 和 statusline cache_hit 段测量实际命中率
- **优化**：把缓存断点放在稳定内容（系统提示词、指令）上，不要放在可变数据（查询、时间戳）上

更多详情请参考 [Anthropic 提示词缓存文档](https://platform.claude.com/docs/en/docs/build-with-claude/prompt-caching)。
