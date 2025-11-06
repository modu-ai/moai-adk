---
title: 9-feedback 命令指南
description: 学习如何使用 Alfred 的 9-feedback 命令快速创建 GitHub Issues 和收集团队反馈
---

# 9-feedback 命令指南

`/alfred:9-feedback` 命令是 MoAI-ADK 反馈收集系统的核心工具，帮助开发者快速创建 GitHub Issues、收集团队意见、追踪问题解决，确保项目持续改进。

## 命令概览

### 基本语法
```bash
/alfred:9-feedback
```

### 命令目的
- 快速创建 GitHub Issues
- 收集团队反馈和建议
- 追踪问题和功能请求
- 优化项目改进流程
- 促进团队协作沟通

### 触发的代理
- **feedback-collector**：主导反馈收集
- **issue-manager**：GitHub Issue 管理
- **team-coordinator**：团队协调
- **quality-analyzer**：反馈质量分析

---

## 反馈类型详解

### 1. Bug 报告 (Bug Report)

#### 何时使用
- 发现系统错误或异常行为
- 功能不符合预期
- 性能问题或崩溃
- 安全漏洞或风险

#### 交互流程
```bash
# 启动反馈流程
/alfred:9-feedback
```

**Alfred 会询问**：

```
🐛 您想创建什么类型的反馈？
[ ] Bug Report - 发现了问题或错误
[ ] Feature Request - 想要新功能
[ ] Improvement - 改进现有功能
[ ] Question/Discussion - 需要团队讨论
```

**选择 Bug Report 后**：

```
📋 请输入问题标题（简洁明确）：
示例：登录页面在移动设备上崩溃

📄 请详细描述问题（可选，按 Enter 跳过）：
示例：
当使用 iPhone 13 访问登录页面时，点击登录按钮会导致应用崩溃。
复现步骤：
1. 打开应用
2. 导航到登录页面
3. 输入正确邮箱和密码
4. 点击登录按钮
5. 应用崩溃

预期行为：应该成功登录并跳转到主页
实际行为：应用崩溃，显示"应用程序已停止"错误

环境信息：
- 设备：iPhone 13 (iOS 17.2)
- 应用版本：v2.1.0
- 网络环境：WiFi 4G
- 发生时间：2025-01-15 14:30

🎯 请选择问题优先级：
[ ] 🔴 Critical - 系统崩溃、数据丢失、安全问题
[ ] 🟠 High - 主要功能无法使用
[✓] 🟡 Medium - 一般功能受影响（默认）
[ ] 🟢 Low - 轻微问题，可以稍后处理
```

#### 自动生成的 GitHub Issue
```markdown
# 🐛 [BUG] 登录页面在移动设备上崩溃

## 问题概述
用户在 iPhone 13 上访问登录页面时，点击登录按钮会导致应用崩溃。

## 环境信息
- **设备**: iPhone 13 (iOS 17.2)
- **应用版本**: v2.1.0
- **网络环境**: WiFi 4G
- **发生时间**: 2025-01-15 14:30

## 复现步骤
1. 打开应用
2. 导航到登录页面
3. 输入正确邮箱和密码
4. 点击登录按钮
5. 应用崩溃

## 预期行为
应该成功登录并跳转到主页

## 实际行为
应用崩溃，显示"应用程序已停止"错误

## 错误信息
```
应用程序已停止
进程 com.example.app 已停止
```

## 附件
- [崩溃日志](link-to-crash-logs)
- [设备截图](link-to-screenshots)

## 优先级
🟡 Medium - 影响用户体验但不影响核心功能

## 相关链接
- `@SPEC:USER-AUTH-001`: 用户认证系统
- `@TEST:USER-AUTH-001`: 认证测试套件
- `@CODE:USER-AUTH-001:API`: 认证 API 实现

## 标签
bug, mobile, crash, login, priority-medium

---

📄 此 Issue 由 Alfred 自动生成，基于用户提供的信息
🔗 相关代码：`@CODE:USER-AUTH-001:API`
🐛 问题追踪：需要 iOS 团队调查和修复
```

### 2. 功能请求 (Feature Request)

#### 何时使用
- 需要新功能或增强
- 想改进用户体验
- 需要新的集成或扩展
- 有产品改进建议

#### 交互流程
```bash
# 启动反馈流程
/alfred:9-feedback
```

**选择 Feature Request 后**：

```
💡 请输入功能请求标题（简洁明确）：
示例：添加暗黑模式支持

📄 请详细描述功能需求（可选，按 Enter 跳过）：
示例：
希望应用支持暗黑模式，提高用户在不同光线条件下的使用体验。

功能需求：
1. 在设置中提供暗黑模式开关
2. 自动跟随系统暗黑模式设置
3. 记住用户的模式偏好
4. 在所有界面中正确应用暗黑主题

用户价值：
- 减少眼部疲劳
- 改善夜间使用体验
- 提升应用现代化感
- 符合系统设计规范

技术考虑：
- 需要创建暗黑主题样式文件
- 实现主题切换逻辑
- 确保所有组件都支持两种主题
- 测试在不同设备上的显示效果

🎯 请选择功能优先级：
[ ] 🔴 Critical - 核心功能，必须立即实现
[ ] 🟠 High - 重要功能，近期实现
[✓] 🟡 Medium - 有价值的功能，计划中实现
[ ] 🟢 Low - 锦上添花的功能，有时间再实现
```

#### 自动生成的 GitHub Issue
```markdown
# ✨ [FEATURE] 添加暗黑模式支持

## 功能概述
为应用添加暗黑模式支持，提供更好的用户体验和视觉舒适度。

## 用户价值
- ✅ 减少眼部疲劳，特别是在低光环境下
- ✅ 改善夜间使用体验
- ✅ 提升应用现代化感和专业度
- ✅ 符合现代设计规范和用户期望

## 功能需求
### 核心功能
1. **主题切换控件**
   - 在设置中提供暗黑模式开关
   - 支持手动切换模式

2. **自动适配**
   - 自动跟随系统暗黑模式设置
   - 在系统模式变化时自动切换

3. **用户偏好**
   - 记住用户的主题选择
   - 应用重启后保持用户设置

4. **全面覆盖**
   - 所有界面和组件都支持两种主题
   - 确保色彩对比度符合可访问性标准

### 扩展功能
- 定时自动切换（如日落时自动切换到暗黑模式）
- 多种暗黑主题选择（如深灰、深蓝等）
- 桌面端和移动端一致性

## 技术实现方案
### 前端实现
- 创建暗黑主题样式变量
- 实现 CSS 变量系统
- 添加主题切换组件
- 确保组件库支持主题切换

### 状态管理
- 在用户设置中存储主题偏好
- 实现主题状态管理逻辑
- 支持实时主题切换

### 设计考虑
- 确保颜色对比度符合 WCAG 2.1 AA 标准
- 保持品牌一致性
- 测试在各种设备和屏幕上的显示效果

## 验收标准
- [ ] 用户可以在设置中手动切换主题
- [ ] 应用自动跟随系统主题设置
- [ ] 主题偏好在应用重启后保持
- [ ] 所有界面都正确应用暗黑主题
- [ ] 颜色对比度符合可访问性标准
- [ ] 主题切换动画流畅自然
- [ ] 在不同设备上显示一致

## 优先级
🟡 Medium - 有价值的用户体验改进，计划在下个季度实现

## 技术复杂度
**预估工作量**: 2-3 周
**复杂度**: 中等
**依赖**: 需要设计师提供暗黑主题设计稿

## 相关链接
- `@SPEC:UI-THEME-001`: 主题系统设计
- 设计稿链接: [Figma 暗黑主题设计](link-to-design)
- 可访问性指南: [WCAG 2.1 对比度标准](https://www.w3.org/TR/WCAG21/)

## 标签
feature, enhancement, ui, theme, dark-mode, accessibility

---

💡 此 Issue 由 Alfred 自动生成
🎨 需要设计团队提供暗黑主题设计稿
📱 需要在所有平台上测试主题显示效果
♿ 确保符合可访问性标准
```

### 3. 改进建议 (Improvement)

#### 何时使用
- 发现现有功能的改进机会
- 有性能优化建议
- 用户体验改进想法
- 代码质量或架构改进建议

#### 交互流程
```bash
# 启动反馈流程
/alfred:9-feedback
```

**选择 Improvement 后**：

```
⚡ 请输入改进建议标题（简洁明确）：
示例：优化搜索功能性能

📄 请详细描述改进建议（可选，按 Enter 跳过）：
示例：
当前搜索功能在大数据量情况下响应较慢，需要进行性能优化。

当前问题：
- 搜索 1000+ 条记录时响应时间超过 3 秒
- 没有搜索结果缓存机制
- 缺少搜索历史和建议功能
- 搜索结果显示不直观

改进方案：
1. **性能优化**
   - 实现搜索结果缓存
   - 添加数据库索引优化
   - 实现搜索结果分页
   - 使用 Elasticsearch 或类似搜索引擎

2. **用户体验**
   - 添加搜索历史记录
   - 实现智能搜索建议
   - 优化搜索结果显示
   - 添加搜索过滤器

3. **技术实现**
   - 引入 Redis 缓存
   - 实现异步搜索
   - 优化数据库查询
   - 添加搜索分析

预期效果：
- 搜索响应时间 < 500ms
- 用户体验明显提升
- 系统负载降低
- 功能更加完善

🎯 请选择改进优先级：
[ ] 🔴 Critical - 严重影响用户使用
[ ] 🟠 High - 重要改进，近期处理
[✓] 🟡 Medium - 有价值的改进，计划中处理
[ ] 🟢 Low nice-to-have 改进，有时间再实现
```

#### 自动生成的 GitHub Issue
```markdown
# ⚡ [IMPROVEMENT] 优化搜索功能性能

## 当前问题
搜索功能在大数据量情况下性能较差，用户体验有待提升。

### 具体问题
- **响应时间慢**: 搜索 1000+ 条记录时响应时间超过 3 秒
- **无缓存机制**: 每次搜索都重新查询数据库
- **缺少用户功能**: 没有搜索历史和建议功能
- **显示不直观**: 搜索结果展示不够清晰

## 改进方案

### 1. 性能优化
- **搜索结果缓存**: 使用 Redis 缓存热门搜索结果
- **数据库优化**: 添加适当的索引和查询优化
- **分页实现**: 实现搜索结果分页，减少单次查询量
- **搜索引擎**: 考虑使用 Elasticsearch 或类似搜索引擎

### 2. 用户体验改进
- **搜索历史**: 记录用户搜索历史，方便重复搜索
- **智能建议**: 基于历史数据提供搜索建议
- **结果优化**: 改进搜索结果的展示和排序
- **过滤功能**: 添加搜索过滤器，提高搜索精确度

### 3. 技术实现
- **缓存层**: 引入 Redis 作为搜索结果缓存
- **异步处理**: 实现异步搜索，提升响应速度
- **查询优化**: 优化 SQL 查询，添加必要索引
- **分析监控**: 添加搜索性能监控和分析

## 预期效果
- ✅ 搜索响应时间 < 500ms (当前 > 3000ms)
- ✅ 用户体验明显提升
- ✅ 系统负载降低 40%
- ✅ 功能更加完善和易用

## 技术方案详情

### 缓存策略
```python
# 搜索缓存设计
cache_config = {
    "popular_searches": {"ttl": 3600, "key_pattern": "search:popular:*"},
    "user_searches": {"ttl": 1800, "key_pattern": "search:user:{user_id}:*"},
    "result_cache": {"ttl": 600, "key_pattern": "search:result:{query_hash}"}
}
```

### 数据库优化
```sql
-- 添加搜索相关索引
CREATE INDEX CONCURRENTLY idx_products_search_vector
ON products USING gin(to_tsvector('english', name || ' ' || description));

CREATE INDEX CONCURRENTLY idx_products_name_trgm
ON products USING gin(name gin_trgm_ops);

CREATE INDEX CONCURRENTLY idx_products_category_active
ON products(category, is_active) WHERE is_active = true;
```

## 实施计划

### 第一阶段：性能优化 (1 周)
- [ ] 引入 Redis 缓存
- [ ] 优化数据库查询
- [ ] 实现搜索结果分页
- [ ] 性能测试和基准测试

### 第二阶段：功能增强 (1 周)
- [ ] 实现搜索历史记录
- [ ] 添加搜索建议功能
- [ ] 优化搜索结果显示
- [ ] 添加搜索过滤器

### 第三阶段：分析和监控 (0.5 周)
- [ ] 添加搜索分析功能
- [ ] 实现性能监控
- [ ] 用户行为分析
- [ ] 效果评估和优化

## 优先级
🟡 Medium - 有价值的性能和用户体验改进

## 技术复杂度
**预估工作量**: 2.5 周
**复杂度**: 中等偏高
**风险评估**: 低风险

## 相关链接
- `@SPEC:SEARCH-001`: 搜索功能规格说明
- `@CODE:SEARCH-001:API`: 搜索 API 实现
- 性能基准测试报告: [link-to-performance-report]

## 标签
improvement, performance, optimization, search, cache, user-experience

---

⚡ 此 Issue 由 Alfred 自动生成
⚙️ 建议由后端团队负责性能优化部分
🎨 建议由前端团队负责用户体验改进部分
📊 建议实施后进行性能基准测试
```

### 4. 问题讨论 (Question/Discussion)

#### 何时使用
- 需要团队讨论的技术决策
- 有架构或设计疑问
- 需要收集团队意见
- 有不确定的实现方案

#### 交互流程
```bash
# 启动反馈流程
/alfred:9-feedback
```

**选择 Question/Discussion 后**：

```
❓ 请输入讨论主题（简洁明确）：
示例：关于微服务架构的技术选型讨论

📄 请详细描述讨论内容（可选，按 Enter 跳过）：
示例：
我们正在讨论新功能的微服务架构设计，需要团队就技术选型达成共识。

背景信息：
- 计划开发用户管理、订单管理、支付系统三个模块
- 考虑使用微服务架构提高系统可扩展性
- 需要确定具体的技术栈和实现方案

讨论选项：
1. **技术栈选项 A**
   - 语言：Python + FastAPI
   - 数据库：PostgreSQL + Redis
   - 消息队列：RabbitMQ
   - 容器化：Docker + Kubernetes

   优点：
   - 团队熟悉 Python 生态
   - FastAPI 性能优秀
   - 丰富的第三方库

   缺点：
   - Python 性能相对较低
   - 内存使用较多

2. **技术栈选项 B**
   - 语言：Go + Gin
   - 数据库：PostgreSQL + Redis
   - 消息队列：Apache Kafka
   - 容器化：Docker + Kubernetes

   优点：
   - Go 性能优秀，并发处理能力强
   - 内存占用低
   - 编译型语言，部署简单

   缺点：
   - 团队需要学习 Go 语言
   - 生态相对较小

3. **技术栈选项 C**
   - 语言：Java + Spring Boot
   - 数据库：PostgreSQL + Redis
   - 消息队列：Apache Kafka
   - 容器化：Docker + Kubernetes

   优点：
   - Java 生态成熟，企业级应用广泛
   - Spring Boot 功能丰富
   - 性能优秀，稳定性好

   缺点：
   - 开发效率相对较低
   - 内存占用较高

讨论重点：
- 团队技能匹配度
- 性能要求 vs 开发效率
- 维护成本和技术支持
- 长期发展考虑

🎯 请选择讨论优先级：
[ ] 🔴 Critical - 关键决策，需要尽快确定
[ ] 🟠 High - 重要决策，本周内确定
[✓] 🟡 Medium - 一般讨论，可适当安排
[ ] 🟢 Low - 一般性讨论，有空时讨论
```

#### 自动生成的 GitHub Issue
```markdown
# ❓ [DISCUSSION] 关于微服务架构的技术选型讨论

## 讨论背景
我们正在讨论新功能的微服务架构设计，需要团队就技术选型达成共识。

## 项目背景
- **目标系统**: 用户管理、订单管理、支付系统
- **架构模式**: 微服务架构
- **业务需求**: 高可用、可扩展、易维护

## 技术选项对比

### 选项 A: Python + FastAPI
**技术栈**:
- 语言：Python 3.11+
- Web 框架：FastAPI
- 数据库：PostgreSQL + Redis
- 消息队列：RabbitMQ
- 容器化：Docker + Kubernetes

**优势**:
✅ 团队熟悉 Python 生态，学习成本低
✅ FastAPI 性能优秀，API 文档自动生成
✅ 丰富的第三方库，开发效率高
✅ 异步支持良好，适合 I/O 密集型应用

**挑战**:
<span class="material-icons">warning</span> Python 性能相对较低，CPU 密集型任务可能有瓶颈
<span class="material-icons">warning</span> 内存使用相对较高，需要优化资源配置
<span class="material-icons">warning</span> GIL 限制，真正的并行计算受限

### 选项 B: Go + Gin
**技术栈**:
- 语言：Go 1.21+
- Web 框架：Gin
- 数据库：PostgreSQL + Redis
- 消息队列：Apache Kafka
- 容器化：Docker + Kubernetes

**优势**:
✅ Go 性能优秀，并发处理能力强
✅ 内存占用低，资源利用率高
✅ 编译型语言，部署简单
✅ 原生并发支持，无需额外框架

**挑战**:
<span class="material-icons">warning</span> 团队需要学习 Go 语言，初期开发效率可能较低
<span class="material-icons">warning</span> 生态相对较小，第三方库选择有限
<span class="material-icons">warning</span> Web 框架功能相对简单，需要更多自定义开发

### 选项 C: Java + Spring Boot
**技术栈**:
- 语言：Java 17+
- Web 框架：Spring Boot 3.x
- 数据库：PostgreSQL + Redis
- 消息队列：Apache Kafka
- 容器化：Docker + Kubernetes

**优势**:
✅ Java 生态成熟，企业级应用广泛
✅ Spring Boot 功能丰富，开发效率高
✅ 性能优秀，稳定性好
✅ 社区支持好，技术资料丰富

**挑战**:
<span class="material-icons">warning</span> 开发效率相对较低，代码量较大
<span class="material-icons">warning</span> 内存占用较高，资源配置需要优化
<span class="material-icons">warning</span> 启动时间较长，微服务实例启动慢

## 决策考虑因素

### 技术因素
- **性能要求**: 需要支持高并发和低延迟
- **开发效率**: 需要平衡开发速度和代码质量
- **团队技能**: 考虑现有团队的技术栈熟悉度
- **维护成本**: 长期维护和运营成本

### 业务因素
- **时间压力**: 项目交付时间要求
- **功能复杂度**: 业务逻辑的复杂程度
- **扩展性要求**: 未来的扩展需求
- **合规要求**: 是否有特定的合规要求

### 运营因素
- **监控能力**: 日志、监控、告警方案
- **部署便利性**: CI/CD 和部署流程
- **故障处理**: 故障排查和恢复能力
- **成本控制**: 基础设施和运营成本

## 讨论重点

1. **团队技能匹配度**
   - 现有团队的技术栈背景
   - 学习新技术的时间和成本
   - 招聘和技术培训考虑

2. **性能 vs 开发效率**
   - 当前阶段更看重哪个方面
   - 性能瓶颈预估和应对方案
   - 开发速度对项目时间的影响

3. **长期发展考虑**
   - 技术栈的发展趋势
   - 社区支持和生态系统
   - 企业级应用需求

4. **风险评估**
   - 每种技术选型的风险点
   - 风险缓解策略
   - 应急方案考虑

## 时间安排
**讨论开始**: 2025-01-15 15:00
**预期决策**: 2025-01-17 18:00
**目标**: 团队就技术选型达成一致意见

## 参与人员
- @tech-lead: 技术负责人
- @backend-team: 后端开发团队
- @devops-team: 运维团队
- @product-manager: 产品经理

## 预期产出
- [ ] 技术选型决策文档
- [ ] 详细的技术方案设计
- [ ] 技术风险评估报告
- [ ] 实施计划和里程碑

## 标签
discussion, architecture, technology-stack, microservices, decision

---

❓ 此 Issue 由 Alfred 自动生成
🤝 需要团队成员积极参与讨论
📋 请在截止日期前完成讨论并做出决策
📄 所有讨论内容将记录在决策文档中
```

---

## 高级功能

### 1. 批量反馈收集

#### 语法
```bash
# 从文件批量创建 Issues
/alfred:9-feedback --batch-file feedback_list.txt

# 从特定目录收集反馈
/alfred:9-feedback --scan-directory ./feedbacks/

# 从会议记录提取问题
/alfred:9-feedback --from-meeting-notes meeting_notes.md
```

#### 文件格式示例
```markdown
# feedback_list.txt

## Bug Report
标题: 支付页面加载缓慢
描述: 用户反映支付页面加载时间超过 5 秒，影响用户体验
优先级: High

## Feature Request
标题: 支持批量导入用户数据
描述: 希望支持通过 CSV 文件批量导入用户数据，提高管理员效率
优先级: Medium

## Improvement
标题: 优化邮件发送性能
描述: 当前邮件发送使用同步方式，影响系统响应速度
优先级: Low
```

### 2. 模板化反馈

#### 语法
```bash
# 使用预定义模板
/alfred:9-feedback --template bug-report

# 自定义模板
/alfred:9-feedback --template custom_template.md

# 创建新模板
/alfred:9-feedback --create-template performance_issue
```

#### 预定义模板示例
```yaml
# bug-report 模板
模板类型: bug-report
必填字段:
  - 标题
  - 描述
  - 复现步骤
  - 环境信息
  - 优先级

可选字段:
  - 错误日志
  - 截图/附件
  - 相关代码
  - 预期行为

# feature-request 模板
模板类型: feature-request
必填字段:
  - 标题
  - 功能描述
  - 用户价值
  - 验收标准
  - 优先级

可选字段:
  - 技术方案
  - 实现建议
  - 相关资源
  - 时间安排
```

### 3. 智能分类和路由

#### 自动分类
```python
def classify_feedback(feedback_text):
    """智能分类反馈"""

    # 使用关键词分析进行分类
    bug_keywords = ["错误", "崩溃", "异常", "问题", "失败", "不工作"]
    feature_keywords = ["新增", "功能", "需求", "建议", "希望"]
    improvement_keywords = ["优化", "改进", "性能", "体验", "建议"]

    classification = {
        "type": "unknown",
        "priority": "medium",
        "component": "unknown",
        "urgency": "normal"
    }

    # 分析文本确定类型
    if any(keyword in feedback_text for keyword in bug_keywords):
        classification["type"] = "bug"
    elif any(keyword in feedback_text for keyword in feature_keywords):
        classification["type"] = "feature"
    elif any(keyword in feedback_text for keyword in improvement_keywords):
        classification["type"] = "improvement"

    # 分析优先级
    urgency_keywords = ["紧急", "严重", "崩溃", "无法使用"]
    if any(keyword in feedback_text for keyword in urgency_keywords):
        classification["priority"] = "critical"

    return classification
```

#### 智能路由
```python
def route_feedback(feedback):
    """智能路由反馈给合适的团队"""

    classification = classify_feedback(feedback["content"])

    routing = {
        "bug": {
            "assignees": ["@backend-team", "@qa-team"],
            "labels": ["bug", "triage-needed"],
            "milestone": "bug-fixing"
        },
        "feature": {
            "assignees": ["@product-manager", "@tech-lead"],
            "labels": ["enhancement", "needs-estimation"],
            "milestone": "feature-development"
        },
        "improvement": {
            "assignees": ["@tech-lead"],
            "labels": ["enhancement", "performance"],
            "milestone": "optimization"
        }
    }

    return routing.get(classification["type"], routing["improvement"])
```

### 4. 反馈分析报告

#### 生成统计报告
```bash
# 生成反馈统计报告
/alfred:9-feedback --analytics

# 按时间范围分析
/alfred:9-feedback --analytics --period=30days

# 按团队分析
/alfred:9-feedback --analytics --team=backend
```

#### 分析报告示例
```markdown
# 反馈分析报告

## 统计概览
**时间范围**: 2025-01-01 至 2025-01-15
**总反馈数**: 127 个

## 反馈类型分布
| 类型 | 数量 | 占比 | 趋势 |
|------|------|------|------|
| Bug Report | 68 | 53.5% | ⬆️ 增加 12% |
| Feature Request | 35 | 27.6% | ⬇️ 减少 5% |
| Improvement | 24 | 18.9% | ➡️ 持平 |

## 优先级分布
| 优先级 | 数量 | 占比 |
|--------|------|------|
| Critical | 8 | 6.3% |
| High | 42 | 33.1% |
| Medium | 52 | 40.9% |
| Low | 25 | 19.7% |

## 组件分布
| 组件 | 问题数 | 占比 | 状态 |
|------|--------|------|------|
| 用户认证 | 18 | 14.2% | 12 已修复 |
| 支付系统 | 15 | 11.8% | 8 已修复 |
| 搜索功能 | 12 | 9.4% | 5 已修复 |
| 数据同步 | 10 | 7.9% | 3 已修复 |

## 响应时间分析
| 时间范围 | 平均响应时间 | 目标时间 |
|----------|--------------|----------|
| Critical | 0.5 小时 | 1 小时 |
| High | 2.3 小时 | 4 小时 |
| Medium | 8.5 小时 | 8 小时 |
| Low | 24.7 小时 | 24 小时 |

## 解决率统计
| 时间范围 | 新增问题 | 已解决 | 解决率 |
|----------|----------|--------|--------|
| 本周 | 25 | 23 | 92% |
| 上周 | 32 | 28 | 87.5% |
| 本月 | 127 | 105 | 82.7% |

## 趋势分析
### 积极趋势
- ✅ Bug 报告响应时间缩短 15%
- ✅ Critical 问题解决率提升到 87%
- ✅ 平均修复时间减少 20%

### 需要关注
- <span class="material-icons">warning</span> Feature Request 处理时间增长
- <span class="material-icons">warning</span> Low 优先级问题积压增加
- <span class="material-icons">warning</span> 部分组件问题重复出现

## 改进建议
1. **优化流程**
   - 加强 Feature Request 的评估流程
   - 定期清理 Low 优先级积压问题
   - 建立问题预防机制

2. **团队协作**
   - 加强跨团队沟通
   - 定期召开问题复盘会议
   - 分享最佳实践和解决方案

3. **工具改进**
   - 优化反馈分类算法
   - 增强自动分配功能
   - 改进报告生成模板

## 下一步行动
- [ ] 处理积压的 Low 优先级问题
- [ ] 优化 Feature Request 评估流程
- [ ] 加强问题预防措施
- [ ] 更新反馈处理 SLA
```

---

## 使用示例

### 示例 1：快速 Bug 报告
```bash
# 快速报告发现的问题
/alfred:9-feedback

# 输出：
✅ GitHub Issue #245 创建成功
🐛 标题: [BUG] 登录按钮在移动设备上无响应
🟠 优先级: High
🏷️ 标签: bug, mobile, login, ui, priority-high
🔗 链接: https://github.com/company/repo/issues/245
```

### 示例 2：功能请求
```bash
# 提出新功能建议
/alfred:9-feedback

# 输出：
✅ GitHub Issue #246 创建成功
✨ 标题: [FEATURE] 添加数据导出功能
🟡 优先级: Medium
🏷️ 标签: feature, enhancement, data-export, reporting
🔗 链接: https://github.com/company/repo/issues/246
```

### 示例 3：团队讨论
```bash
# 发起技术讨论
/alfred:9-feedback

# 输出：
✅ GitHub Discussion #123 创建成功
❓ 标题: [DISCUSSION] 关于 API 版本管理策略
🟡 优先级: Medium
🏷️ 标签: discussion, architecture, api, versioning
🔗 链接: https://github.com/company/repo/discussions/123
```

### 示例 4：批量反馈处理
```bash
# 批量处理团队反馈
/alfred:9-feedback --batch-file team_feedback.txt

# 输出：
✅ 批量创建 5 个 Issues
✅ Bug Reports: 3 个
✅ Feature Requests: 2 个
✅ 自动分配给相应团队
✅ 设置适当标签和优先级
📋 详细报告: .moai/reports/batch-feedback-2025-01-15.md
```

---

## 与 GitHub 集成

### 1. GitHub Actions 集成
```yaml
# .github/workflows/feedback-automation.yml
name: 反馈自动化

on:
  issues:
    types: [opened, closed]

jobs:
  auto-classify:
    runs-on: ubuntu-latest
    steps:
    - name: 分类和标记 Issue
      run: |
        echo "Issue ${{ github.event.issue.number }} opened"
        # 这里可以调用 Alfred 的分类逻辑

  auto-assign:
    runs-on: ubuntu-latest
    steps:
    - name: 自动分配 Issue
      run: |
        # 根据内容自动分配给合适的团队
        echo "自动分配 Issue 给相应团队"
```

### 2. 项目管理工具集成
```bash
# 与 Jira 集成
/alfred:9-feedback --create-jira-ticket

# 与 Trello 集成
/alfred:9-feedback --create-trello-card

# 与 Asana 集成
/alfred:9-feedback --create-asana-task
```

### 3. 通知集成
```bash
# 发送 Slack 通知
/alfred:9-feedback --notify-slack

# 发送邮件通知
/alfred:9-feedback --notify-email

# 发送 Microsoft Teams 通知
/alfred:9-feedback --notify-teams
```

---

## 最佳实践

### 1. 编写有效的反馈

#### Bug 报告最佳实践
```bash
# ✅ 好的 Bug 报告
标题: [BUG] 登录页面在 Safari 浏览器上显示异常

描述:
当使用 Safari 浏览器访问登录页面时，页面布局显示异常，登录按钮被遮挡。
详细描述：
- 影响浏览器：Safari 16.x
- 影响设备：MacBook Pro, iPhone
- 影响版本：v2.1.0
- 复现频率：100%

# <span class="material-icons">cancel</span> 不好的 Bug 报告
标题: 登录有问题
描述: 登录不好用
```

#### 功能请求最佳实践
```bash
# ✅ 好的功能请求
标题: [FEATURE] 支持多种语言界面切换

描述:
希望能够支持中文、英文、日文三种语言界面切换，提升国际化用户体验。

用户价值：
- 扩展国际市场
- 提升用户满意度
- 符合国际化趋势

验收标准：
- 支持三种语言界面切换
- 保存用户语言偏好
- 所有界面元素都支持多语言

# <span class="material-icons">cancel</span> 不好的功能请求
标题: 要多语言支持
描述: 希望支持多种语言
```

### 2. 反馈时机

#### 及时反馈
```bash
# 发现问题立即反馈
/alfred:9-feedback

# 不要等待问题积累
```

#### 定期反馈收集
```bash
# 每周团队会议收集反馈
/alfred:9-feedback --team-meeting

# 版本发布后收集用户反馈
/alfred:9-feedback --post-release
```

### 3. 团队协作

#### 反馈循环
```yaml
反馈处理流程:
1. 收集反馈
   - 用户直接反馈
   - 团队内部反馈
   - 测试团队反馈
   - 产品经理反馈

2. 分类和评估
   - 自动分类优先级
   - 技术可行性评估
   - 业务价值评估
   - 资源需求评估

3. 决策和分配
   - 确定处理优先级
   - 分配给相应团队
   - 设定时间期限
   - 制定处理计划

4. 执行和跟踪
   - 按计划执行
   - 定期进度更新
   - 问题解决确认
   - 关闭反馈项

5. 总结和改进
   - 分析处理效果
   - 总结经验教训
   - 优化处理流程
   - 预防类似问题
```

---

## 故障排除

### 常见问题

#### 1. GitHub 连接问题
**症状**: 无法创建 GitHub Issue

**解决方案**:
```bash
# 检查 GitHub CLI 认证
gh auth status

# 重新认证
gh auth login

# 检查仓库权限
gh repo view

# 测试连接
gh issue list
```

#### 2. 分类不准确
**症状**: 自动分类结果不准确

**解决方案**:
```bash
# 手动指定类型
/alfred:9-feedback --type=bug

# 查看分类规则
/alfred:9-feedback --show-classification-rules

# 更新分类规则
/alfred:9-feedback --update-classification
```

#### 3. 模板问题
**症状**: 模板格式不正确

**解决方案**:
```bash
# 验证模板格式
/alfred:9-feedback --validate-template bug-report

# 重置默认模板
/alfred:9-feedback --reset-templates

# 重新下载模板
/alfred:9-feedback --download-templates
```

### 调试技巧

#### 1. 启用详细日志
```bash
# 启用调试模式
export ALFRED_DEBUG=true
/alfred:9-feedback --debug

# 保存调试信息
/alfred:9-feedback --debug --output=debug.log
```

#### 2. 预览功能
```bash
# 预览将要创建的 Issue
/alfred:9-feedback --preview

# 只生成 Markdown，不创建 Issue
/alfred:9-feedback --markdown-only --output=preview.md
```

#### 3. 测试模式
```bash
# 测试模式，不会实际创建 Issue
/alfred:9-feedback --test-mode

# 使用测试仓库
/alfred:9-feedback --test-repo
```

---

## 总结

`/alfred:9-feedback` 命令是 MoAI-ADK 反馈收集系统的核心工具，它能够：

- **快速创建 Issues**：支持 Bug 报告、功能请求、改进建议、团队讨论
- **智能分类路由**：自动分类并分配给合适团队
- **模板化处理**：使用标准模板确保信息完整
- **批量处理支持**：高效处理大量反馈
- **分析报告生成**：提供详细的反馈统计和趋势分析

### 关键要点

1. **及时反馈**：发现问题立即反馈，不要积累
2. **信息完整**：提供足够的上下文和详细信息
3. **明确分类**：使用合适的类型和优先级
4. **团队协作**：促进团队沟通和知识共享
5. **持续改进**：基于反馈数据优化产品和流程

### 下一步

- [学习 0-project 命令](../project/)
- [理解项目管理](../project/init.md)
- [掌握团队协作](../project/team-collaboration.md)
- [查看高级功能](../../advanced/)

通过熟练使用 `/alfred:9-feedback` 命令，您可以建立高效的反馈收集和处理机制，促进团队协作，持续改进产品质量和用户体验。