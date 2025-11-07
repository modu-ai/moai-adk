# Alfred Skills 详细指南

为 Alfred 和子代理设计的 5 个专业技能。

## 概览

| 技能                               | 说明                   | 目标对象   | 版本 |
| ---------------------------------- | ---------------------- | ---------- | ---- |
| **moai-alfred-agent-guide**        | 19 人团队结构, 决策树  | Alfred     | 4.0  |
| **moai-alfred-ask-user-questions** | 用户交互优化           | 所有代理   | 2.1  |
| **moai-alfred-personas**           | 自适应通信             | Alfred     | 3.0  |
| **moai-alfred-best-practices**     | TRUST, TAG, Skill 规则 | 验证       | 5.0  |
| **moai-alfred-context-budget**     | Context window 优化    | Alfred     | 2.5  |

______________________________________________________________________

## 1. moai-alfred-agent-guide

**19 人 AI 团队结构, 选择算法, 协作模式**

### 团队结构

```
Alfred (主管)
├── 10 名核心 Sub-agents
│   ├─ project-manager: 项目初始化
│   ├─ spec-builder: SPEC 编写
│   ├─ implementation-planner: 计划制定
│   ├─ tdd-implementer: TDD 执行
│   ├─ doc-syncer: 文档同步
│   ├─ tag-agent: TAG 管理
│   ├─ git-manager: Git 自动化
│   ├─ trust-checker: 质量验证
│   ├─ quality-gate: 发布准备
│   └─ debug-helper: 错误解决
│
├── 6 名专家 Agents
│   ├─ backend-expert: API/服务器
│   ├─ frontend-expert: UI/状态管理
│   ├─ devops-expert: 部署/CI/CD
│   ├─ ui-ux-expert: 设计/可访问性
│   ├─ security-expert: 安全
│   └─ database-expert: DB 设计
│
└── 2 名内置 Agents
    ├─ Claude Opus/Sonnet: 复杂推理
    └─ Claude Haiku: 轻量级任务
```

### Lead-Specialist Pattern

```python
# Alfred 检测领域关键词
if "database" in spec:
    activate(database_expert)
    # database_expert 与 Alfred 协作
    # Alfred: 整体协调
    # database_expert: DB 专业知识

if "security" in spec:
    activate(security_expert)
    # security_expert 进行安全审查

if "performance" in spec:
    activate(debug_helper)
    # debug_helper 进行性能优化
```

### Master-Clone Pattern

```
大规模任务 (100+ 文件, 5+ 步骤)
    ↓
Master Alfred
├─→ Clone-1: 模块 A (独立执行)
├─→ Clone-2: 模块 B (并行处理)
└─→ Clone-3: 模块 C (同时进行)
    ↓
Master 协调和合并结果
```

### 决策树

```
用户请求
    ↓
Alfred: 领域分析
    ├─ 后端任务? → backend-expert
    ├─ 前端? → frontend-expert
    ├─ 部署? → devops-expert
    ├─ 安全? → security-expert
    ├─ 数据库? → database-expert
    └─ UI/设计? → ui-ux-expert
    ↓
    ├─ 大规模? (100+ 文件) → Master-Clone
    ├─ 领域专业化? → Lead-Specialist
    └─ 一般任务? → Alfred 直接处理
    ↓
激活选定的代理
```

______________________________________________________________________

## 2. moai-alfred-ask-user-questions

**用户交互最佳使用方法**

### 必需规则

```
❌ 禁止表情符号的位置:
- question: "这个设置正确吗?" (❌ "🔧 选择设置?")
- header: "Authentication" (❌ "🔐 认证")
- label: "JWT Token" (❌ "✅ JWT")
- description: "Stateless token" (❌ "🎯 Stateless...")

✅ 允许的位置:
- 响应消息: "✅ 设置完成"
- 说明文本: "💡 提示: JWT 是..."
```

### 结构化问题

```json
{
  "questions": [
    {
      "question": "您想使用哪种认证方式?",
      "header": "Authentication Method",
      "multiSelect": false,
      "options": [
        {
          "label": "JWT",
          "description": "无状态, API 最佳"
        },
        {
          "label": "Session",
          "description": "现有 Web 应用, 保持服务器状态"
        },
        {
          "label": "OAuth 2.0",
          "description": "社交登录, 第三方集成"
        }
      ]
    }
  ]
}
```

### 使用时机

#### ✅ 必须使用

- 需要多种技术选择 (3 个以上)
- 需要架构决策
- 请求模糊
- 影响范围大

#### ❌ 不使用

- 请求明确
- 简单的是/否问题
- 不需要技术决策

### 批处理策略

```python
# 如果需要 5 个以上选项
# → 拆分为多个顺序的 AskUserQuestion 调用

# 示例: 语言选择 → GitHub 设置 → 框架
# 总共 3 个独立的 AskUserQuestion 调用
```

______________________________________________________________________

## 3. moai-alfred-personas

**自适应通信风格**

### 用户级别检测

#### Beginner 级别

```
特征:
- 首次使用 MoAI-ADK
- 不了解 SPEC-First 概念
- 无 TDD 经验

Alfred 的通信:
- 详细的说明
- 分步指南
- 先解释概念
- 提供大量示例
```

#### Intermediate 级别

```
特征:
- 可以使用基本的 Alfred
- 有 SPEC 编写经验
- 了解基本的 TDD

Alfred 的通信:
- 适度的详细程度
- 仅强调核心
- 提供优化提示
- 展示模式
```

#### Expert 级别

```
特征:
- Alfred 熟练用户
- 理解 Master-Clone 模式
- 可以定制需求

Alfred 的通信:
- 简洁的说明
- 技术细节
- 优化策略
- 自定义解决方案
```

### 角色检测信号

```
Beginner 信号:
- "SPEC 是什么?"
- "如何进行 TDD?"
- 请求详细说明

Intermediate 信号:
- 使用 SPEC 关键词
- 功能实现请求
- 架构问题

Expert 信号:
- 大规模迁移
- 自定义代理
- 性能优化
```

______________________________________________________________________

## 4. moai-alfred-best-practices

**TRUST, TAG, Skill 调用规则**

### TRUST 5 强制要求

```
❌ 绝对禁止:
- 编写没有测试的代码
- 以低于 85% 的覆盖率部署
- 忽略安全漏洞
- 忽略可追溯性

✅ 必需要求:
- 严格遵守 RED-GREEN-REFACTOR
- 所有实现都有 @CODE TAG
- 所有测试都有 @TEST TAG
- 所有文档都有 @DOC TAG
```

### TAG 链验证

```
SPEC-001
    ↓
@TEST:SPEC-001:* (最少 1 个)
    ↓
@CODE:SPEC-001:* (最少 1 个)
    ↓
@DOC:SPEC-001:* (最少 1 个)
    ↓
全部完成时 ✅
```

### Skill 调用规则

```python
# ✅ 正确调用
Skill("moai-lang-python")
Skill("moai-domain-backend")

# ❌ 错误调用
Skill("python")  # 错误!
Skill("backend")  # 错误!

# ✅ 必需调用 (验证前)
Skill("moai-foundation-trust")
Skill("moai-foundation-tags")
```

______________________________________________________________________

## 5. moai-alfred-context-budget

**Context Window 优化**

### Context 分配策略

```
总 Context Window: 200,000 tokens

分配:
├── System Prompt: 10,000 tokens (5%)
├── Conversation History: 80,000 tokens (40%)
├── Current Task: 40,000 tokens (20%)
├── Code Files: 50,000 tokens (25%)
└── Reserve: 20,000 tokens (10%)
```

### JIT (Just-In-Time) 加载

```python
# ❌ 一次性加载所有文件
Read("file1.py")
Read("file2.py")
Read("file3.py")
Read("file4.py")

# ✅ 仅在需要时加载
Read("file1.py")  # 仅需要的文件
# ... 执行任务
Read("file2.py")  # 下一个需要的文件
```

### Memory 文件模式

```
.moai/
├── .session-memory.md        # 当前会话状态
├── .plan-summary.md          # 当前计划摘要
└── .progress-snapshot.md     # 进度快照

大小优化:
- 每个文件 < 10KB
- 摘要格式 (不详细)
- 自动清理 (定期)
```

### 清理策略

```
会话结束时:
✅ 归档完成的任务
✅ 删除临时文件
✅ 压缩大型日志文件
✅ 摘要 Memory 文件
```

______________________________________________________________________

## Alfred Skills 集成工作流

```
用户请求
    ↓
Skill("moai-alfred-agent-guide")
├── 确认团队结构
├── 执行决策树
└── 选择需要的代理
    ↓
Skill("moai-alfred-ask-user-questions")
├── 需要澄清? → 用户交互
└── 进行 → 下一步
    ↓
Skill("moai-alfred-personas")
├── 检测用户级别
└── 调整通信
    ↓
Skill("moai-alfred-context-budget")
├── Context 效率
└── Memory 优化
    ↓
Skill("moai-alfred-best-practices")
├── TRUST 5 验证
├── TAG 链确认
└── Skill 调用验证
    ↓
任务执行
    ↓
完成
```

______________________________________________________________________

## Alfred Skills FAQ

### "应该激活哪个代理?"

→ 参考 `Skill("moai-alfred-agent-guide")` 中的决策树

### "Context 不足"

→ 使用 `Skill("moai-alfred-context-budget")` 进行优化

### "用户的请求不明确"

→ 使用 `Skill("moai-alfred-ask-user-questions")` 进行交互

### "如何验证 TRUST 5?"

→ 参考 `Skill("moai-alfred-best-practices")` 的 TRUST 部分

______________________________________________________________________

**下一步**: [Foundation Skills](foundation.md) 或 [Skills 概览](index.md)
