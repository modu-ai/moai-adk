---
title: 参与贡献
weight: 110
draft: false
---

MoAI-ADK 是开源项目，欢迎贡献！MoAI-ADK 自身也是用基于 SPEC 的 3-phase
工作流和 TRUST 5 质量门开发的 —— 贡献流程的质量标准（覆盖率、Lint、
Conventional Commit）沿用同一套标准。

{{< mascot coding >}}

## 快速开始

1. **Fork** 仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写测试（新代码用 TDD，现有代码用特征化测试）
4. 确认所有测试通过：`make test`
5. 确认 Lint 通过：`make lint`
6. 格式化代码：`make fmt`
7. 用 Conventional Commit 消息提交
8. 创建 Pull Request

## 代码质量要求

TRUST 5 框架的 **T**ested / **T**rackable 标准原样适用：

| 项目 | 标准 |
|------|------|
| 测试覆盖率 | **85%** 以上 |
| Lint 错误 | **0** 个 |
| 类型错误 | **0** 个 |
| 提交信息 | Conventional Commits 格式 |

## 提交信息格式

```
<type>(<scope>): <description>

[可选正文]

[可选页脚]
```

### 类型

| 类型 | 说明 |
|------|------|
| `feat` | 新功能 |
| `fix` | 修复错误 |
| `docs` | 文档变更 |
| `style` | 代码格式（无功能变更） |
| `refactor` | 重构（无功能变更） |
| `perf` | 性能改进 |
| `test` | 添加/修改测试 |
| `chore` | 构建/工具变更 |
| `revert` | 回滚之前的提交 |

### 示例

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## 开发环境设置

### 必需工具

- **Go 1.26+** —— 核心开发语言
- **Git** —— 版本管理
- **make** —— 构建命令

### 主要命令

```bash
make build        # 构建项目
make test         # 运行测试
make test-race    # Race condition 检测测试
make lint         # 运行 Linter
make fmt          # 格式化代码
make install      # 本地安装
make clean        # 清理构建产物
```

## Pull Request 指南

### 撰写 PR 时

- 清晰简洁的标题（70 字符以内）
- 变更内容摘要（Summary 部分）
- 测试计划（Test Plan 部分）
- 引用相关 issue（例：`Fixes #123`）

### PR 检查清单

- [ ] 添加/更新测试
- [ ] 所有测试通过（`make test`）
- [ ] Lint 通过（`make lint`）
- [ ] 提交信息符合 Conventional Commits 格式
- [ ] 更新文档（如有必要）

## 社区

- **问题跟踪器**: [GitHub Issues](https://github.com/modu-ai/moai-adk/issues) —— 错误报告、功能请求。如果正在使用 MoAI-ADK，可以用 `/moai feedback` 直接在会话内创建 issue
- **Discord**: [Discord 社区](https://discord.gg/Z7E7Mdc5aN) —— 实时交流、分享技巧
- **官方文档**: [adk.mo.ai.kr](https://adk.mo.ai.kr)

## 许可证

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) —— 可自由使用、修改、分发。
