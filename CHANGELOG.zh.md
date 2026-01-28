# v1.9.0 - Memory MCP、SVG 技能、规则迁移 (2026-01-26)

## 概要

引入跨会话持久内存、全面的 SVG 技能和符合标准的规则系统迁移的次要版本。

**主要功能**:
- **Memory MCP 集成**: 用户偏好和项目上下文的持久存储
- **SVG 技能**: 包含 SVGO 优化模式和最佳实践的综合技能
- **规则迁移**: 从 `.moai/rules/*.yaml` 迁移到 `.claude/rules/*.md`（Claude Code 官方标准）
- **错误修复**: Rank batch sync 显示问题（#300）

**影响**:
- 通过 Memory MCP 启用代理间上下文共享
- 专业的 SVG 创建和优化支持
- 更清洁、符合标准的项目结构
- 准确的批处理同步统计显示

## Breaking Changes

无。所有更改都向后兼容。

## 新增

### Memory MCP 集成

- **feat**: 添加 Memory MCP Server 集成 (99ab5273)
  - Claude Code 会话间的持久内存
  - 用户偏好、项目上下文、学习模式存储
  - 工作流期间代理间上下文共享
  - 配置: `.mcp.json`, `.mcp.windows.json`
  - 新技能: `moai-foundation-memory` (420 行)

### SVG 创建和优化技能

- **feat**: 添加 `moai-tool-svg` 技能 (54c12a85)
  - 基于 W3C SVG 2.0 规范和 SVGO 文档
  - 全面的模块: 基础、样式、优化、动画
  - 12 个可工作的代码示例
  - SVGO 配置模式和最佳实践
  - 总共 3,698 行（SKILL.md: 410、modules: 2,288、examples: 500、reference: 500）

### 语言规则增强

- **feat**: 使用增强的工具信息更新语言规则 (54c12a85)
  - Ruff 配置模式（替换 flake8+isort+pyupgrade）
  - Mypy strict mode 指南
  - 测试框架推荐
  - 更新了 16 个语言文件

## 更改

### CLAUDE.md 优化

- **refactor**: v1.9.0 的大规模清理和模块化 (4134e60d)
  - 将 CLAUDE.md 从 ~60k 减少到 ~30k 字符（符合 40k 限制）
  - 将详细内容移至 `.claude/rules/` 以改善组织
  - 添加 `shell_validator.py` 实用程序以实现跨平台兼容性
  - 增强 CLI 命令（doctor、init、update）
  - 添加 `moai-workflow-thinking` 技能
  - 添加 bug-report.yml 问题模板
  - 影响: 改善可读性、可维护性和 Claude Code 兼容性

### 规则系统迁移

- **feat**: 从 `.moai/rules/*.yaml` 迁移到 `.claude/rules/*.md` (99ab5273)
  - 删除: 6,959 行 YAML 规则
  - 添加: Claude Code 官方 Markdown 规则
  - 结构: `.claude/rules/{core,development,workflow,languages}/`
  - 影响: 标准合规性、更清洁的组织

## 修复

### Rank 命令

- **fix(rank)**: 正确解析 batch sync 的嵌套 API 响应 (#300) (31b504ed)
  - 问题: `moai-adk rank sync` 始终显示 "Submitted: 0"
  - 根本原因: 缺少嵌套的 `data` 字段提取
  - 修复: 在访问字段之前添加 `data = response.get("data", {})`
  - 影响: 准确的提交统计显示

## 安装和更新

```bash
# 更新到最新版本
uv tool update moai-adk

# 更新项目文件夹中的模板
moai update

# 验证版本
moai --version
```

---

# v1.8.13 - Statusline Context Window 修复 (2026-01-26)

## 概要

改进 statusline context window 计算准确性的补丁版本。

**主要修复**:
- 修复 statusline context window 百分比以使用 Claude Code 的预计算值

**影响**:
- Context window 显示现在考虑了自动压缩和输出令牌保留
- 更准确的剩余令牌信息

## 修复

### Statusline Context Window 计算

- **fix(statusline)**: 使用 Claude Code 的预计算 context 百分比 (2dacecb7)
  - 优先级 1: 使用 Claude Code 的 `used_percentage`/`remaining_percentage`（最准确）
  - 优先级 2: 从 `current_usage` 令牌计算（回退）
  - 优先级 3: 没有数据时返回 0%（会话开始）
  - 确保启用自动压缩或保留输出令牌时的准确性
  - 文件: `src/moai_adk/statusline/main.py`

## 安装和更新

```bash
# 更新到最新版本
uv tool update moai-adk

# 更新项目模板
moai update

# 验证版本
moai --version
```

---

# v1.8.12 - Hook Format Update & Login Command (2026-01-26)

## 概要

包含 Claude Code hook format 兼容性修复和 UX 改进的补丁版本。

**主要更改**:
- 修复 Claude Code settings.json hook format（新的基于匹配器的结构）
- 将 `moai rank register` 重命名为 `moai rank login`（更直观）
- settings.json 现在在更新时始终被覆盖；使用 settings.local.json 进行自定义

**影响**:
- MoAI Rank hooks 现在可以在最新的 Claude Code 中工作
- `moai rank login` 是新的主要命令（register 仍可作为别名使用）
- 用户自定义保存在 settings.local.json 中

## Breaking Changes

无。`moai rank register` 仍可作为隐藏别名使用。
