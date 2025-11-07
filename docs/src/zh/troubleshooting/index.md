# 问题排查指南

使用MoAI-ADK过程中遇到问题的解决方法。

## 🎯 按问题类型查找解决方案

### 安装与初始化问题

- [安装错误](https://adk.mo.ai.kr/troubleshooting/installation)
- [初始化失败](https://adk.mo.ai.kr/troubleshooting/initialization)
- [环境配置](https://adk.mo.ai.kr/troubleshooting/environment)

### Alfred命令问题

- [无法识别命令](https://adk.mo.ai.kr/troubleshooting/command-not-found)
- [SPEC生成失败](https://adk.mo.ai.kr/troubleshooting/spec-creation)
- [TDD周期错误](https://adk.mo.ai.kr/troubleshooting/tdd-errors)

### 开发与构建问题

- [测试失败](https://adk.mo.ai.kr/troubleshooting/test-failures)
- [依赖项错误](https://adk.mo.ai.kr/troubleshooting/dependency-errors)
- [构建错误](https://adk.mo.ai.kr/troubleshooting/build-errors)

### Git与部署问题

- [Git冲突](https://adk.mo.ai.kr/troubleshooting/git-conflicts)
- [部署失败](https://adk.mo.ai.kr/troubleshooting/deployment-errors)
- [CI/CD问题](https://adk.mo.ai.kr/troubleshooting/cicd-issues)

______________________________________________________________________

## ❓ 常见问题 (FAQ)

### 基本使用

**Q: 如何开始使用MoAI-ADK?**
A: 请参考[快速入门指南](../getting-started/quick-start.md)。3分钟内即可完成基本配置。

**Q: 什么是SPEC-First?**
A: [基本概念](../getting-started/concepts.md)中有详细说明。简单来说，就是在编写代码之前先编写规范的方式。

**Q: Alfred的作用是什么?**
A: 请查看[Alfred工作流](../guides/alfred/index.md)。Alfred是协调19个AI专家团队的超级代理。

### TDD相关

**Q: TDD的RED-GREEN-REFACTOR是什么?**
A: [TDD指南](../guides/tdd/index.md)详细说明了各个阶段。

**Q: 测试覆盖率应该达到多少?**
A: MoAI-ADK建议**85%以上的测试覆盖率**。

### TAG系统

**Q: 为什么需要@TAG系统?**
A: 通过[TAG系统](../guides/specs/tags.md)可以连接SPEC、TEST、CODE、DOC，提供完整的可追溯性。

______________________________________________________________________

## 🚨 常见错误消息

### "Command not found: /alfred:1-plan"

**原因**: Claude Code无法识别Alfred命令

**解决方案**:

```bash
# 1. 重启Claude Code
exit
claude

# 2. 检查目录
ls .claude/commands/

# 3. 刷新设置
/alfred:0-project
```

### "SPEC file not found"

**原因**: SPEC文件未在正确位置生成

**解决方案**:

```bash
# 检查项目状态
moai-adk doctor

# 检查.moai/目录权限
ls -la .moai/

# 重新初始化
rm -rf .moai
/alfred:0-project
```

### "Test coverage below 85%"

**原因**: 测试覆盖率不足

**解决方案**:

```bash
# 检查当前覆盖率
pytest --cov=src tests/

# 添加缺失的测试
# 在tests/test_*.py中添加测试用例

# 重新运行
pytest --cov=src tests/
```

______________________________________________________________________

## 🔧 系统诊断

### 运行诊断工具

```bash
# 检查整体系统状态
moai-adk doctor

# 详细输出
moai-adk doctor --verbose
```

### 检查项目

- Python版本与依赖项
- Git设置与权限
- .moai/目录结构
- .claude/配置文件
- 必需工具安装状态

______________________________________________________________________

## 💬 获取更多帮助

### 社区

- [GitHub Discussions](https://github.com/modu-ai/moai-adk/discussions) - 提问并分享想法
- [Issue Tracker](https://github.com/modu-ai/moai-adk/issues) - 报告Bug

### 文档

- [在线文档](https://adk.mo.ai.kr) - 最新信息
- [本地文档](../index.md) - 离线参考

### 反馈

```bash
# 报告问题 (自动创建GitHub Issue)
/alfred:9-feedback
```

______________________________________________________________________

**这对您有帮助吗?** 欢迎[提出更多问题](https://github.com/modu-ai/moai-adk/discussions)！
