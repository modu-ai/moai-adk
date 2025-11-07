# Foundation Skills 详细指南

所有 MoAI-ADK 项目的基础 4 个核心技能。

## 概览

| 技能                      | 说明                | 版本 | 调用方式                                  |
| ------------------------- | ------------------- | ---- | ----------------------------------------- |
| **moai-foundation-trust** | TRUST 5 原则验证    | 5.0  | `Skill("moai-foundation-trust")`          |
| **moai-foundation-tags**  | TAG 系统 (可追溯性) | 3.2  | `Skill("moai-foundation-tags")`           |
| **moai-alfred-workflow**  | 四步工作流          | 4.0  | `Skill("moai-alfred-workflow")`           |
| **moai-alfred-ask-user**  | 用户交互            | 2.1  | `Skill("moai-alfred-ask-user-questions")` |

______________________________________________________________________

## 1. moai-foundation-trust

**TRUST 5 原则验证和应用**

### 五大原则

#### 🧪 T - Test First

**测试驱动开发**

```
需求
    ↓
编写测试 (RED)
    ↓
最小实现 (GREEN)
    ↓
代码改进 (REFACTOR)
    ↓
验证测试覆盖率 85%+
```

**验证标准**:

- 覆盖率 85% 以上
- 所有测试通过
- 包含边缘案例

#### 📖 R - Readable

**易读的代码**

```python
# ❌ 难读的代码
def f(x):
    return sum([i*2 for i in x if i>0])

# ✅ 易读的代码
def double_positive_numbers(numbers):
    """返回正数翻倍的列表"""
    return [num * 2 for num in numbers if num > 0]
```

**验证项目**:

- MyPy/type-checking 通过
- 遵守命名规则 (camelCase/snake_case)
- 函数长度 50 行以下
- 复杂度 10 以下

#### 🎯 U - Unified

**统一的结构**

```
项目结构:
src/
  ├── models/       # 数据模型
  ├── api/         # API 端点
  ├── services/    # 业务逻辑
  ├── utils/       # 工具
  └── config.py    # 配置

tests/
  ├── unit/        # 单元测试
  ├── integration/ # 集成测试
  └── e2e/        # E2E 测试
```

**验证项目**:

- 项目结构一致性
- 遵守命名规则
- 导入结构一致性

#### 🔒 S - Secured

**保证安全性**

```python
# ❌ 安全风险
user = User.query.filter_by(
    email=request.args.get('email')  # SQL 注入风险!
).first()

# ✅ 安全的代码
from sqlalchemy import text
user = User.query.filter_by(
    email=request.args.get('email')  # SQLAlchemy ORM 自动转义
).first()
```

**验证项目**:

- 无 OWASP Top 10 漏洞
- 依赖安全检查 (Snyk, safety)
- 输入验证和转义
- 加密存储

#### 🏷️ T - Trackable

**完整的可追溯性**

```
SPEC-001 (需求)
    ↓
@TEST:SPEC-001 (测试)
    ↓
@CODE:SPEC-001 (实现)
    ↓
@DOC:SPEC-001 (文档)
    ↓
全部相互引用
```

**验证项目**:

- 所有实现都包含 TAG
- TAG 链完成 (SPEC→TEST→CODE→DOC)
- 无孤立 TAG

### TRUST 验证自动化

```bash
# 执行 TRUST 5 验证
Skill("moai-foundation-trust")

# 验证结果
✅ Test First: 92% 覆盖率
✅ Readable: MyPy 通过
✅ Unified: 遵守结构
✅ Secured: 安全检查通过
✅ Trackable: TAG 完成

🎯 TRUST 5: PASS ✅
```

______________________________________________________________________

## 2. moai-foundation-tags

**TAG 系统完整指南**

### TAG 语法

#### SPEC TAG

```
SPEC-001: 第一个规范
SPEC-002: 第二个规范
```

#### TEST TAG

```
@TEST:SPEC-001:login_feature
@TEST:SPEC-001:password_validation
```

#### CODE TAG

```
@CODE:SPEC-001:register_user
@CODE:SPEC-001:validate_email
```

#### DOC TAG

```
@DOC:SPEC-001:api_documentation
@DOC:SPEC-001:deployment_guide
```

### TAG 链示例

```python
# @CODE:SPEC-001:user_registration
def register_user(email: str, password: str) -> User:
    """
    用户注册 @CODE:SPEC-001:register_user

    @TEST:SPEC-001:test_register_success 参考
    """
    # @CODE:SPEC-001:validate_email
    if not is_valid_email(email):
        raise ValueError("Invalid email")

    # @CODE:SPEC-001:hash_password
    hashed = hash_password(password)

    # @CODE:SPEC-001:create_user
    user = User(email=email, password_hash=hashed)
    db.session.add(user)
    db.session.commit()

    return user

# @TEST:SPEC-001:test_register_success
def test_register_success():
    """@TEST:SPEC-001:test_register_success"""
    user = register_user("test@example.com", "password123")
    assert user.email == "test@example.com"
    # @TEST:SPEC-001:verify_user_created 验证
    assert user.id is not None
```

### TAG 验证规则

| 规则             | 说明                               |
| ---------------- | ---------------------------------- |
| **禁止重复**     | 同一 TAG 在多个文件中存在则错误    |
| **禁止孤立 TAG** | 无对应 SPEC 的 TAG 删除            |
| **链完成**       | SPEC→TEST→CODE→DOC 全部连接        |
| **明确识别**     | TAG 应该唯一且可追溯               |

### TAG 扫描

```bash
# 查询 TAG 状态
moai-adk status --spec SPEC-001

# TAG 验证
/alfred:3-sync auto SPEC-001

# TAG 去重
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

______________________________________________________________________

## 3. moai-alfred-workflow

**四步 Alfred 工作流**

### Phase 1: 意图理解 (Intent Understanding)

```
用户请求 → 明确性评估
    ├─ 明确: 进入 Phase 2
    └─ 不明确: AskUserQuestion → 用户响应 → 进入 Phase 2
```

### Phase 2: 计划制定 (Plan Creation)

```
调用 Plan Agent
    ↓
├─ 任务分解 (Decomposition)
├─ 依赖关系分析 (Dependency Analysis)
├─ 识别并行化机会 (Parallelization)
├─ 明确文件列表 (File List)
└─ 时间估计 (Time Estimation)
    ↓
用户批准 (AskUserQuestion)
    ↓
初始化 TodoWrite
```

### Phase 3: 任务执行 (Execution)

```
RED Phase
├─ 编写测试
└─ 确认全部失败

GREEN Phase
├─ 最小实现
└─ 确认全部通过

REFACTOR Phase
├─ 代码改进
└─ 保持测试
```

### Phase 4: 报告和提交 (Report & Commit)

```
任务完成
    ↓
├─ 文档生成 (根据生成配置)
├─ Git 提交 (自动)
├─ PR 创建 (团队模式)
└─ 清理
```

______________________________________________________________________

## 4. moai-alfred-ask-user-questions

**用户交互优化**

### AskUserQuestion 使用方法

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
          "description": "无状态, REST API 最佳"
        },
        {
          "label": "OAuth 2.0",
          "description": "第三方服务集成"
        },
        {
          "label": "Session",
          "description": "现有 Web 应用"
        }
      ]
    }
  ]
}
```

### 最佳使用场景

| 场景           | 是否使用 | 说明             |
| -------------- | -------- | ---------------- |
| 明确的请求     | ❌       | 直接进行         |
| 模糊的请求     | ✅       | 澄清             |
| 技术决策       | ✅       | 提供选项         |
| 架构选择       | ✅       | 说明权衡         |
| 影响范围确认   | ✅       | 事先通知         |

### 规则

```
- ❌ 禁止表情符号 (question, header, label, description)
- ✅ 最多 4 个选项 (5 个以上时多次调用)
- ✅ 结构化格式 (header + options)
- ✅ 清晰的说明 (每个选项)
```

______________________________________________________________________

## Foundation Skills 集成示例

```
用户请求
    ↓
Skill("moai-alfred-workflow") - 应用四步工作流
    ↓
Phase 1: 意图理解
    └─→ Skill("moai-alfred-ask-user-questions") - 澄清
    ↓
Phase 2: 计划制定
    └─→ 初始化 TodoWrite
    ↓
Phase 3: 任务执行 (TDD)
    └─→ Skill("moai-foundation-trust") - TRUST 5 验证
    └─→ Skill("moai-foundation-tags") - 添加 TAG
    ↓
Phase 4: 报告和提交
    └─→ Git 提交 (自动)
    ↓
完成!
```

______________________________________________________________________

## Foundation Skills FAQ

### "TRUST 5 严格吗?"

→ **非常严格**。覆盖率低于 85% 无法部署。

### "需要始终添加所有 TAG 吗?"

→ **是的, 为了可追溯性是必需的**。孤立的 TAG 会被自动删除。

### "可以跳过四步工作流吗?"

→ **不可以, 始终遵循**。Phase 1 是必需的。

______________________________________________________________________

**下一步**: [Languages Skills](languages.md) 或 [Alfred Skills](alfred.md)
