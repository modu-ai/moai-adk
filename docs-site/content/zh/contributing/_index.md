---
title: 贡献
weight: 110
draft: false
---

MoAI-ADK 是一个开源项目，我们欢迎您的贡献！本指南说明如何为项目做出贡献。

## 快速开始

1. **Fork** 仓库
2. 创建功能分支：`git checkout -b feature/my-feature`
3. 编写测试（新代码使用 TDD，现有代码使用特性化测试）
4. 确认所有测试通过：`make test`
5. 确认 linting 通过：`make lint`
6. 格式化代码：`make fmt`
7. 使用 Conventional Commit 消息提交
8. 创建 Pull Request

## 代码质量要求

| 项目 | 标准 |
|------|------|
| 测试覆盖率 | **85%** 或以上 |
| Lint 错误 | **0** 个 |
| 类型错误 | **0** 个 |
| 提交消息 | Conventional Commits 格式 |

## 提交消息格式

```
<type>(<scope>): <description>

[可选本文]

[可选页脚]
```

### 类型

| 类型 | 描述 |
|------|------|
| `feat` | 新功能 |
| `fix` | 错误修复 |
| `docs` | 文档更改 |
| `style` | 代码格式（无功能更改） |
| `refactor` | 重构（无功能更改） |
| `perf` | 性能改进 |
| `test` | 测试添加/修改 |
| `chore` | 构建/工具更改 |
| `revert` | 还原先前提交 |

### 示例

```
feat(template): add SessionEnd hook to settings.json generator
fix(cli): prevent race condition in hook execution
test(settings): add TestEnsureGlobalSettingsEnv test cases
docs(readme): update agent count and statistics
```

## 开发环境设置

### 必需工具

- **Go 1.26+** — 核心开发语言
- **Git** — 版本管理
- **make** — 构建命令

### 主要命令

```bash
make build        # 构建项目
make test         # 运行测试
make test-race    # 检测竞态条件
make lint         # 运行 linter
make fmt          # 格式化代码
make install      # 本地安装
make clean        # 清理构建产物
```

## Pull Request 指南

### 创建 PR 时

- 清晰简洁的标题（70 字以内）
- 变更内容摘要（Summary 部分）
- 测试计划（Test Plan 部分）
- 相关问题参考（例如：`Fixes #123`）

### PR 检查清单

- [ ] 添加/更新测试
- [ ] 所有测试通过（`make test`）
- [ ] Linting 通过（`make lint`）
- [ ] 提交消息遵循 Conventional Commits 格式
- [ ] 更新文档（如需要）

## 社区

- **问题跟踪**：[GitHub Issues](https://github.com/modu-ai/moai-adk/issues) — 错误报告、功能请求
- **官方文档**：[adk.mo.ai.kr](https://adk.mo.ai.kr)

## 许可证

[Apache License 2.0](https://github.com/modu-ai/moai-adk/blob/main/LICENSE) — 自由使用、修改和分发。
