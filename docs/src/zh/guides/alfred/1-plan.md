---
title: 1-plan 命令指南
description: 学习如何使用 Alfred 的 1-plan 命令进行需求分析和规格说明编写
---

# 1-plan 命令指南

`/alfred:1-plan` 命令是 MoAI-ADK 需求分析阶段的核心工具，负责将用户的模糊需求转换为清晰、可执行的规格说明（SPEC）。

## 命令概览

### 基本语法
```bash
/alfred:1-plan "功能描述"
```

### 命令目的
- 分析用户需求
- 创建结构化 SPEC 文档
- 设计验收标准
- 评估技术风险
- 制定实现计划

### 触发的代理
- **spec-builder**：主导 SPEC 编写
- **project-manager**：项目上下文分析
- **domain-experts**：领域专业知识支持

---

## 工作流程详解

### 1. 需求理解阶段

#### 用户输入分析
Alfred 首先分析用户的输入，提取关键信息：

```python
def analyze_user_input(user_input):
    analysis = {
        "core_intent": extract_core_intent(user_input),
        "domain": identify_domain(user_input),
        "scope": determine_scope(user_input),
        "complexity": assess_complexity(user_input),
        "dependencies": identify_dependencies(user_input),
        "constraints": extract_constraints(user_input)
    }
    return analysis
```

#### 上下文信息收集
Alfred 会收集项目上下文信息：

```yaml
上下文信息:
项目配置:
  - 项目类型: web 应用
  - 技术栈: Python + FastAPI
  - 团队规模: 3-5 人
  - 开发阶段: 开发中

历史信息:
  - 相关 SPEC: USER-001, USER-002
  - 类似功能: 用户管理模块
  - 技术债务: 需要重构的用户服务

约束条件:
  - 时间限制: 2 周内完成
  - 性能要求: 响应时间 < 200ms
  - 兼容性: 支持现有 API 版本
```

### 2. 需求澄清阶段

#### 互动式提问
如果需求不够明确，Alfred 会提出澄清问题：

```bash
Alfred: 为了更好地理解您的需求，我需要澄清几个问题：

Q1: 当您说"用户管理"时，具体包括哪些功能？
☑ 用户注册和登录
☑ 用户信息管理
☑ 权限管理
☐ 用户活动追踪

Q2: 对于用户认证，您希望支持哪些方式？
☑ 邮箱 + 密码
☑ 手机号 + 验证码
☐ 第三方登录（Google、GitHub）
☐ SSO 单点登录

Q3: 性能要求是什么？
☑ 高优先级：响应时间 < 200ms
☑ 中等优先级：支持 1000 并发用户
☐ 低优先级：暂无特殊要求
```

#### 需求边界确定
Alfred 帮助明确功能边界：

```yaml
需求边界分析:
包含功能:
  - 用户注册（邮箱验证）
  - 用户登录（JWT 认证）
  - 密码重置
  - 基本用户信息管理

不包含功能:
  - 第三方登录集成
  - 角色权限系统
  - 用户活动日志
  - 社交功能

未来迭代:
  - 权限管理（v2.0）
  - 用户分析（v2.1）
  - 社交功能（v3.0）
```

### 3. SPEC 编写阶段

#### SPEC 结构生成
Alfred 自动生成标准的 SPEC 文档结构：

```yaml
---
id: USER-AUTH-001
version: 0.1.0
status: draft
priority: high
created: 2025-01-15
updated: 2025-01-15
author: @developer
reviewer: @team-lead
tags: [authentication, user-management, security]
---

# `@SPEC:USER-AUTH-001: 用户认证系统

## 概述
实现完整的用户认证系统，包括用户注册、登录、密码重置等功能，确保系统安全性和用户体验。

## Ubiquitous Requirements
- 系统必须提供用户注册功能
- 系统必须提供用户登录功能
- 系统必须提供密码重置功能
- 系统必须使用 JWT 令牌进行身份验证

## Event-driven Requirements
- 当用户提供有效邮箱和密码时，系统必须创建用户账户并发送验证邮件
- 当用户点击验证邮件链接时，系统必须激活用户账户
- 当用户提供正确凭证时，系统必须发放 JWT 访问令牌
- 当用户提供注册邮箱时，系统必须发送密码重置链接
- 当用户访问受保护资源时，系统必须验证 JWT 令牌有效性

## State-driven Requirements
- 当用户处于未验证状态时，系统必须限制访问受保护资源
- 当用户处于已验证状态时，系统必须允许访问用户相关资源
- 当令牌处于过期状态时，系统必须要求重新登录

## Optional Requirements
- 如果用户有头像文件，系统可以显示用户头像
- 如果配置了邮件服务，系统可以发送通知邮件

## Constraints
- 密码长度必须在 8-128 字符之间
- 密码必须包含大小写字母、数字和特殊字符
- JWT 令牌过期时间不得超过 24 小时
- API 响应时间不应该超过 200 毫秒
- 邮箱验证链接有效期必须是 24 小时

## 验收标准
### 功能验收
- [ ] 用户可以使用邮箱和密码注册账户
- [ ] 注册后必须验证邮箱才能登录
- [ ] 用户可以使用正确凭证登录
- [ ] 登录成功后获得 JWT 令牌
- [ ] 用户可以重置忘记的密码
- [ ] JWT 令牌可以用于访问受保护资源

### 安全验收
- [ ] 密码使用 bcrypt 加密存储
- [ ] JWT 令牌使用强密钥签名
- [ ] API 有适当的速率限制
- [ ] 输入数据经过验证和清理
- [ ] 敏感操作有日志记录

### 性能验收
- [ ] 登录 API 响应时间 < 200ms
- [ ] 注册 API 响应时间 < 500ms
- [ ] 系统支持 1000 并发用户
- [ ] 数据库查询经过优化

## 依赖关系
### 前置条件
- 邮件服务配置完成
- 数据库用户表已创建
- JWT 密钥已生成

### 后置影响
- 用户管理模块需要适配
- API 网关需要更新认证规则
- 前端需要实现登录界面

## 风险评估
### 技术风险
- **中等风险**: JWT 令牌管理复杂性
  - 缓解措施: 使用成熟的 JWT 库
- **低风险**: 邮件服务依赖
  - 缓解措施: 实现多个邮件服务商备选

### 业务风险
- **高风险**: 安全漏洞可能导致用户数据泄露
  - 缓解措施: 严格遵循安全最佳实践，定期安全审计
- **中等风险**: 用户体验可能受邮件验证影响
  - 缓解措施: 优化邮件发送速度，提供清晰的指导

### 缓解措施
- 实施全面的安全测试
- 定期进行安全审计
- 提供详细的操作日志
- 实施监控和告警机制
```

#### EARS 语法应用
Alfred 自动应用 EARS 语法确保需求清晰：

```yaml
EARS 语法示例:

## Ubiquitous Requirements (普遍需求)
- 系统必须提供用户认证功能
- 系统必须支持密码重置

## Event-driven Requirements (事件驱动需求)
- 当用户提供有效凭证时，系统必须发放访问令牌
- 当令牌过期时，系统必须要求重新认证

## State-driven Requirements (状态驱动需求)
- 当用户处于未验证状态时，系统必须限制功能访问
- 当用户处于已验证状态时，系统必须允许完整访问

## Optional Requirements (可选需求)
- 如果支持第三方登录，系统可以集成 OAuth 提供商

## Constraints (约束)
- 密码强度必须满足安全要求
- API 响应时间不应该超过性能阈值
```

### 4. 验证与确认阶段

#### SPEC 质量检查
Alfred 自动检查 SPEC 质量：

```python
def validate_spec_quality(spec):
    checks = {
        "completeness": check_completeness(spec),
        "clarity": check_clarity(spec),
        "testability": check_testability(spec),
        "consistency": check_consistency(spec),
        "feasibility": check_feasibility(spec)
    }
    return checks
```

#### 验收标准生成
Alfred 自动生成可测试的验收标准：

```yaml
验收标准示例:
Gherkin 格式测试场景:

Feature: 用户认证
  作为一个用户
  我希望能够安全地注册和登录
  以便访问系统的功能

Scenario: 成功的用户注册
  Given 我是一个新用户
  When 我在注册页面提供有效的邮箱和密码
  Then 我应该收到验证邮件
  And 我的账户应该处于未验证状态

Scenario: 成功的用户登录
  Given 我是一个已验证的用户
  When 我提供正确的邮箱和密码
  Then 我应该收到 JWT 访问令牌
  And 令牌应该包含我的用户信息
```

### 5. 输出生成阶段

#### 主要输出
1. **SPEC 文档**：完整的规格说明文档
2. **实现计划**：技术实现建议
3. **风险评估**：潜在风险和缓解措施
4. **验收标准**：可测试的验收条件

#### 辅助输出
1. **Git 分支**：自动创建功能分支
2. **任务清单**：详细的实现任务列表
3. **资源清单**：需要的工具和依赖

---

## 使用示例

### 示例 1：Web API 功能

#### 用户输入
```bash
/alfred:1-plan "创建产品管理 API，支持产品的增删改查操作"
```

#### Alfred 处理过程
1. **需求分析**：识别 CRUD 操作需求
2. **上下文收集**：检查现有产品相关代码
3. **需求澄清**：询问产品属性、权限要求等
4. **SPEC 编写**：生成完整的 API 规格说明
5. **输出结果**：

```yaml
输出结果:
<span class="material-icons">check_circle</span> SPEC ID: PRODUCT-001
<span class="material-icons">check_circle</span> 文件位置: .moai/specs/SPEC-PRODUCT-001/spec.md
<span class="material-icons">check_circle</span> 功能分支: feature/SPEC-PRODUCT-001
<span class="material-icons">check_circle</span> 实现计划: 3个阶段，预计 5 天完成
<span class="material-icons">check_circle</span> 风险评估: 低风险，现有技术栈完全支持

主要功能:
- POST /products - 创建产品
- GET /products - 获取产品列表
- GET /products/{id} - 获取单个产品
- PUT /products/{id} - 更新产品
- DELETE /products/{id} - 删除产品

验收标准:
- [ ] 所有 API 端点正常工作
- [ ] 输入验证和安全检查到位
- [ ] API 响应时间 < 200ms
- [ ] 完整的错误处理机制
```

### 示例 2：数据处理功能

#### 用户输入
```bash
/alfred:1-plan "实现用户行为数据分析功能，生成用户活跃度报告"
```

#### Alfred 处理过程
1. **领域识别**：数据分析和报告生成
2. **专家激活**：data-science-expert 自动参与
3. **技术分析**：评估数据处理需求
4. **实现方案**：推荐适合的技术栈

```yaml
输出结果:
<span class="material-icons">check_circle</span> SPEC ID: ANALYTICS-001
<span class="material-icons">check_circle</span> 专家参与: data-science-expert
<span class="material-icons">check_circle</span> 技术推荐: pandas + matplotlib + Redis 缓存
<span class="material-icons">check_circle</span> 实现复杂度: 中等
<span class="material-icons">check_circle</span> 预计工期: 2 周

核心功能:
- 用户行为数据收集
- 数据清洗和预处理
- 活跃度指标计算
- 报告生成和导出
- 实时数据更新

技术要求:
- 数据处理性能优化
- 大数据量处理能力
- 准确的统计算法
- 清晰的可视化展示
```

### 示例 3：集成功能

#### 用户输入
```bash
/alfred:1-plan "集成第三方支付系统，支持信用卡和移动支付"
```

#### Alfred 处理过程
1. **外部依赖识别**：第三方支付服务
2. **安全要求评估**：PCI DSS 合规性
3. **专家激活**：security-expert, devops-expert 参与
4. **风险评估**：高安全风险，需要特别处理

```yaml
输出结果:
<span class="material-icons">check_circle</span> SPEC ID: PAYMENT-001
<span class="material-icons">check_circle</span> 安全级别: 高
<span class="material-icons">check_circle</span> 专家参与: security-expert, devops-expert
<span class="material-icons">check_circle</span> 合规要求: PCI DSS Level 1
<span class="material-icons">check_circle</span> 风险评估: 高风险，需要额外安全措施

关键考虑:
- 支付数据加密
- 安全的密钥管理
- 审计日志记录
- 异常处理机制
- 合规性验证

实施建议:
- 使用成熟的支付网关 SDK
- 实施额外的安全层
- 建立监控和告警
- 定期安全审计
```

---

## 高级功能

### 1. 批量需求处理

#### 语法
```bash
/alfred:1-plan "功能1描述" "功能2描述" "功能3描述"
```

#### 处理方式
Alfred 会：
1. 分析功能之间的关联性
2. 确定实现优先级
3. 创建多个 SPEC 或合并 SPEC
4. 制定整体实现计划

### 2. 从文件导入需求

#### 语法
```bash
/alfred:1-plan --file requirements.txt
```

#### 文件格式
```markdown
# requirements.txt
用户管理系统
- 用户注册和登录
- 用户信息管理
- 权限管理

订单管理系统
- 订单创建
- 订单查询
- 订单状态更新
```

### 3. 更新现有 SPEC

#### 语法
```bash
/alfred:1-plan --update SPEC-001 "新的需求描述"
```

#### 处理流程
1. 读取现有 SPEC
2. 分析变更内容
3. 更新相关部分
4. 版本控制
5. 影响分析

---

## 最佳实践

### 1. 编写有效的需求描述

#### <span class="material-icons">check_circle</span> 好的实践
```bash
# 具体明确
/alfred:1-plan "创建用户认证 API，支持邮箱注册、JWT 令牌、密码重置"

# 包含验收标准
/alfred:1-plan "实现产品搜索功能，支持关键词搜索、分类筛选、价格区间过滤"

# 考虑边界条件
/alfred:1-plan "开发文件上传功能，支持 10MB 以内文件，图片自动压缩，病毒扫描"
```

#### <span class="material-icons">cancel</span> 避免的做法
```bash
# 过于模糊
/alfred:1-plan "做个登录功能"

# 缺少上下文
/alfred:1-plan "处理数据"

# 功能过多
/alfred:1-plan "完整的电商系统包括用户管理、商品管理、订单管理、支付、物流等"
```

### 2. 提供充分的上下文

#### 在对话中提供背景信息
```bash
# 在执行 1-plan 之前提供上下文
"我们正在开发一个在线教育平台，主要面向大学生。现有用户系统使用 Django 开发，现在需要添加实时聊天功能以提高学生和老师的互动体验。"

# 然后执行命令
/alfred:1-plan "实现实时聊天系统，支持一对一聊天、群组聊天、文件分享"
```

#### 使用项目标签
```bash
/alfred:1-plan "添加用户个人资料管理功能" --context="社交媒体应用" --priority="high"
```

### 3. 渐进式需求定义

#### 分阶段定义复杂功能
```bash
# 第一阶段：核心功能
/alfred:1-plan "实现基础聊天功能，支持文本消息和表情"

# 第二阶段：扩展功能
/alfred:1-plan "添加聊天室功能，支持多人群聊"

# 第三阶段：高级功能
/alfred:1-plan "集成文件分享和语音消息"
```

### 4. 利用反馈循环

#### 确认理解
```bash
用户: /alfred:1-plan "实现用户管理系统"

Alfred: 我理解您需要实现用户管理系统，包括：
- 用户注册和登录
- 用户信息管理
- 权限管理

这些是否准确？还有什么需要补充的吗？

用户: 是的，但还需要支持第三方登录

Alfred: 好的，我会添加第三方登录功能。您希望支持哪些平台？
```

---

## 故障排除

### 常见问题

#### 1. Alfred 无法理解需求
**症状**：生成的 SPEC 与预期不符

**解决方案**：
```bash
# 提供更具体的描述
/alfred:1-plan "详细描述功能，包括具体场景和验收条件"

# 使用示例说明
/alfred:1-plan "实现类似 Twitter 的微博客功能，用户可以发布 280 字符以内的消息"

# 分解复杂需求
/alfred:1-plan "第一部分：基础消息发布功能"
```

#### 2. SPEC 过于复杂
**症状**：生成的 SPEC 包含过多功能

**解决方案**：
```bash
# 明确范围
/alfred:1-plan "实现最小可行产品版本" --scope="mvp"

# 设置优先级
/alfred:1-plan "核心功能" --priority="high" --exclude="advanced-features"

# 分阶段实现
/alfred:1-plan "第一阶段：基础功能"
```

#### 3. 技术可行性问题
**症状**：Alfred 提示技术风险过高

**解决方案**：
```bash
# 请求简化方案
/alfred:1-plan "简化版功能，降低技术复杂度"

# 请求替代方案
/alfred:1-plan "功能描述" --request-alternatives

# 提供技术约束
/alfred:1-plan "功能描述" --tech-stack="Python, PostgreSQL, Redis"
```

### 调试技巧

#### 1. 查看中间结果
```bash
# 检查需求分析结果
/alfred:1-plan "描述" --show-analysis

# 查看生成的 SPEC 草稿
/alfred:1-plan "描述" --draft-only

# 保存调试信息
/alfred:1-plan "描述" --debug --output=debug.log
```

#### 2. 分步执行
```bash
# 只执行需求分析
/alfred:1-plan "描述" --analysis-only

# 只生成 SPEC 结构
/alfred:1-plan "描述" --structure-only

# 跳过风险评估
/alfred:1-plan "描述" --skip-risk-assessment
```

---

## 与其他命令的集成

### 与 2-run 的集成
```bash
# 完整的工作流程
/alfred:1-plan "用户认证功能"
# Alfred 生成 SPEC-AUTH-001

/alfred:2-run AUTH-001
# Alfred 基于 SPEC 实现 TDD
```

### 与 3-sync 的集成
```bash
# 更新需求
/alfred:1-plan --update AUTH-001 "添加双因子认证"

# 同步变更
/alfred:3-sync
# Alfred 更新相关文档和代码
```

### 与 0-project 的集成
```bash
# 项目初始化时定义核心需求
/alfred:0-project
# 配置项目基本信息

/alfred:1-plan "核心业务功能"
# 定义项目的主要功能需求
```

---

## 总结

`/alfred:1-plan` 命令是 MoAI-ADK 需求分析阶段的核心工具，它能够：

- **理解用户需求**：通过自然语言处理和互动式澄清
- **生成专业 SPEC**：使用 EARS 语法和标准模板
- **评估技术可行性**：分析风险和依赖
- **制定实现计划**：提供清晰的执行路线图

### 关键要点

1. **明确具体**：提供清晰、具体的功能描述
2. **充分上下文**：给出足够的背景信息
3. **渐进式开发**：将复杂功能分解为小步骤
4. **及时反馈**：与 Alfred 互动确认理解
5. **版本控制**：使用版本管理需求的变更

### 下一步

- [学习 2-run 命令](2-run.md)
- [理解 TDD 流程](../tdd/)
- [掌握 SPEC 编写](../specs/)
- [查看项目配置](../project/)

通过熟练使用 `/alfred:1-plan` 命令，您可以确保开发团队始终在构建正确的产品，满足用户的真实需求。