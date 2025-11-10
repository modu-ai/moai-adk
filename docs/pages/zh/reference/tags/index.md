# TAG系统完整参考

TAG系统是MoAI-ADK可追溯性系统的核心。

## 目的

通过CODE-FIRST原则连接SPEC、TEST、CODE和DOC，确保**完全可追溯性**。

```
SPEC-001（需求）
    ↓
@TEST:SPEC-001（测试）
    ↓
@CODE:SPEC-001（实现）
    ↓
@DOC:SPEC-001（文档）
    ↓
交叉引用（完全可追溯性）
```

## TAG类型

| TAG         | 位置         | 用途        | 示例              |
| ----------- | ------------ | ----------- | ----------------- |
| **SPEC-ID** | .moai/specs/ | 需求        | SPEC-001          |
| **@TEST**   | tests/       | 测试代码    | @TEST:SPEC-001:\* |
| **@CODE**   | src/         | 实现代码    | @CODE:SPEC-001:\* |
| **@DOC**    | docs/        | 文档        | @DOC:SPEC-001:\*  |

## TAG编写规则

### SPEC TAG

```
SPEC-001: 第一个规范
SPEC-002: 第二个规范
SPEC-N: 第N个规范
```

### @TEST TAG

```python
# @TEST:SPEC-001:login_success
def test_login_success():
    pass

# @TEST:SPEC-001:login_failure
def test_login_failure():
    pass
```

### @CODE TAG

```python
# @CODE:SPEC-001:register_user
def register_user(email, password):
    pass

# @CODE:SPEC-001:validate_email
def validate_email(email):
    pass
```

### @DOC TAG

```markdown
# API文档 @DOC:SPEC-001:api

这是SPEC-001的API文档。
```

## TAG验证规则

| 规则       | 说明                              | 违反时 |
| ---------- | --------------------------------- | ------ |
| **唯一性** | 相同的TAG不能重复                 | 错误   |
| **完整性** | SPEC→TEST→CODE→DOC都必须存在     | 警告   |
| **一致性** | TAG格式一致性                     | 错误   |
| **可追溯性** | 可以交叉引用                      | 警告   |

## TAG扫描和验证

```bash
# 查询TAG状态
moai-adk status

# 特定SPEC TAG详细查询
moai-adk status --spec SPEC-001

# 执行TAG验证
/alfred:3-sync auto SPEC-001

# 删除重复TAG
/alfred:tag-dedup --dry-run
/alfred:tag-dedup --apply --backup
```

## <span class="material-icons">library_books</span> 详细指南

- **[TAG类型](types.md)** - 每种TAG类型的详细说明
- **[可追溯性系统](traceability.md)** - TAG链和完整性验证

______________________________________________________________________

**下一步**: [TAG类型](types.md)或[可追溯性系统](traceability.md)



