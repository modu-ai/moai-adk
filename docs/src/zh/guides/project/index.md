---
title: 项目管理指南
description: MoAI-ADK 项目初始化、配置、部署完整指南
status: stable
---

# 项目管理指南

学习如何管理 MoAI-ADK 项目的完整生命周期。涵盖从初始化到配置、部署的所有阶段。

## 🎯 项目管理 3 阶段

### [1. 项目初始化](init.md)
- 使用 `moai-adk init` 命令创建新项目
- 选择项目模板
- 自动生成必要的文件结构

**主要生成项**:
- `.moai/config.json` - 项目元数据
- `.claude/` - Claude Code 配置 (agents, commands, skills, hooks)
- `pyproject.toml` - Python 项目配置
- `pytest.ini` - 测试配置

### [2. 配置管理](config.md)
- `.moai/config.json` 详细配置
- 语言和本地化设置
- 开发环境定制
- Hook 和 Agent 配置

**核心配置**:
- 项目元数据 (名称、版本、描述)
- 语言和领域设置
- Git 工作流配置
- 报告生成策略

### [3. 部署策略](deploy.md)
- 本地开发环境配置
- Docker 容器化
- 云平台部署 (Vercel, Railway, AWS)
- CI/CD 管道构建
- 监控和日志

## 📊 项目结构

```
my-awesome-project/
├── .moai/              # MoAI-ADK 元数据
│   ├── config.json     # 项目配置
│   ├── docs/           # 自动生成文档
│   └── reports/        # 分析和报告
├── .claude/            # Claude Code 配置
│   ├── agents/         # Sub-agent 定制
│   ├── commands/       # 斜杠命令
│   ├── skills/         # 项目特定 Skill
│   └── hooks/          # 自动化 Hooks
├── src/                # 源代码
├── tests/              # 测试代码
├── docs/               # 项目文档
└── pyproject.toml      # Python 项目配置
```

## 🔄 Alfred 集成

项目管理与 Alfred SuperAgent 完美集成:

- `/alfred:0-project` - 项目配置优化
- `/alfred:1-plan` - 需求 SPEC 编写
- `/alfred:2-run` - TDD 实现
- `/alfred:3-sync` - 文档同步和部署

[Alfred 工作流完整指南](../alfred/index.md)

## 📋 检查清单

项目配置时应确认的事项:

- [ ] 项目初始化完成 (`moai-adk init`)
- [ ] `.moai/config.json` 审查和定制
- [ ] Git 工作流配置确认
- [ ] 开发环境配置完成
- [ ] CI/CD 管道配置 (可选)
- [ ] 部署策略决定

## 🚀 下一步

- [项目初始化: init.md](init.md)
- [配置管理: config.md](config.md)
- [部署策略: deploy.md](deploy.md)
- [Alfred 0-project: 配置优化](../alfred/index.md)

---

**了解更多**: 项目管理是 MoAI-ADK 工作流的基础。正确的配置开始会大大提高开发生产力。
