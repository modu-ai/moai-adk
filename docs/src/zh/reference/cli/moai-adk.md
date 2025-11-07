# moai-adk 命令完整参考

`moai-adk` 命令行工具的所有命令和选项。

## 命令结构

```
moai-adk <command> [options] [arguments]
```

## 全局选项

所有 `moai-adk` 命令均可使用:

| 选项               | 说明           |
| ------------------ | -------------- |
| `--help`, `-h`     | 显示帮助信息   |
| `--version`, `-v`  | 显示版本信息   |
| `--verbose`, `-vv` | 详细输出模式   |
| `--no-color`       | 无颜色输出     |

## 命令列表

### 1. moai-adk init

**项目初始化及模板注入**

#### 语法

```bash
moai-adk init [路径] [选项]
```

#### 参数

| 参数   | 说明         | 默认值       |
| ------ | ------------ | ------------ |
| `路径` | 项目路径     | 当前目录     |

#### 选项

```bash
--language LANG, -l LANG    选择项目语言 (ko/en/ja/zh)
--mode MODE                 开发模式 (solo/team/org)
--with-mcp SERVER          添加 MCP 服务器 (context7, figma, playwright)
--mcp-auto                 自动安装所有推荐的 MCP 服务器
--force, -f                覆盖现有配置
--skip-git                 跳过 Git 初始化
```

#### 示例

```bash
# 创建新项目
moai-adk init my-project

# 初始化当前目录
moai-adk init .

# 包含 MCP 服务器
moai-adk init . --with-mcp context7 --with-mcp figma

# 自动安装所有 MCP
moai-adk init . --mcp-auto

# 强制重新初始化
moai-adk init . --force
```

#### 生成的文件

```
项目/
├── .moai/
│   ├── config.json        # 项目配置
│   ├── specs/            # SPEC 文档
│   ├── docs/             # 生成的文档
│   ├── reports/          # 分析报告
│   └── scripts/          # 实用工具脚本
├── .claude/
│   ├── settings.json     # Claude Code 配置
│   ├── commands/         # Alfred 命令
│   ├── agents/          # Sub-agent 模板
│   └── mcp.json         # MCP 配置
└── CLAUDE.md            # 项目指令
```

______________________________________________________________________

### 2. moai-adk doctor

**系统环境诊断**

#### 语法

```bash
moai-adk doctor [选项]
```

#### 选项

```bash
--verbose, -vv       详细诊断信息
--fix                尝试自动修复
--export FILE       将结果导出到文件
```

#### 诊断项目

- ✅ Python 版本 (需要 3.13+)
- ✅ uv 包管理器
- ✅ Git 仓库状态
- ✅ `.moai/` 目录结构
- ✅ `.claude/` 资源
- ✅ Claude Code 可访问性
- ✅ Python 依赖项
- ✅ 磁盘空间

#### 示例

```bash
# 基本诊断
moai-adk doctor

# 详细诊断
moai-adk doctor -vv

# 保存结果到文件
moai-adk doctor --export .moai/reports/doctor.txt

# 自动修复
moai-adk doctor --fix
```

______________________________________________________________________

### 3. moai-adk status

**查询项目状态**

#### 语法

```bash
moai-adk status [选项]
```

#### 选项

```bash
--json                 JSON 格式输出
--compact, -c          仅显示简要摘要
--spec ID              查询特定 SPEC 详情
```

#### 显示信息

- 📋 SPEC 进度状态 (完成/进行中/待办)
- 🏷️ TAG 统计 (@SPEC/@TEST/@CODE/@DOC)
- 📝 最近提交
- 📅 最后同步时间
- 🔄 Git 分支状态

#### 示例

```bash
# 完整状态
moai-adk status

# JSON 格式
moai-adk status --json

# 简要摘要
moai-adk status --compact

# 特定 SPEC 详情
moai-adk status --spec SPEC-001
```

______________________________________________________________________

### 4. moai-adk backup

**创建项目备份**

#### 语法

```bash
moai-adk backup [选项]
```

#### 选项

```bash
--target DIR           备份位置 (默认: .moai-backups/)
--include-git          包含 Git 历史
--compress, -z         压缩格式 (tar.gz)
--restore FILE         恢复备份
```

#### 备份内容

- `.moai/` 完整目录
- `.claude/` 资源
- `CLAUDE.md` 项目指令
- `pyproject.toml` / `requirements.txt`

#### 示例

```bash
# 基本备份
moai-adk backup

# 压缩备份
moai-adk backup --compress

# 包含 Git 历史
moai-adk backup --include-git

# 恢复备份
moai-adk backup --restore .moai-backups/20250115_143000/

# 自定义位置
moai-adk backup --target ~/backups/moai/
```

______________________________________________________________________

### 5. moai-adk update

**包和模板同步 (最重要的命令)**

#### 语法

```bash
moai-adk update [选项]
```

#### 选项

```bash
--check                仅检查是否有更新
--dry-run              预览变更
--skip-backup          跳过备份 (不推荐)
--force                强制更新
--from VERSION         从特定版本更新
```

#### 执行流程

1. **版本检查**: 从 PyPI 检查最新版本
2. **创建备份**: 安全备份当前状态
3. **模板同步**: 合并新模板和现有配置
4. **验证**: 完整性检查
5. **完成**: 变更摘要

#### 示例

```bash
# 常规更新
moai-adk update

# 预览
moai-adk update --dry-run

# 强制更新
moai-adk update --force

# 仅检查版本
moai-adk update --check
```

______________________________________________________________________

## 选项组合

### 高级使用案例

```bash
# 带详细日志的初始化
moai-adk init . --verbose --with-mcp context7

# 自动修复和详细报告
moai-adk doctor --fix --export .moai/reports/doctor.md

# 导出完整状态为 JSON
moai-adk status --json > status.json

# 创建压缩备份并测试恢复
moai-adk backup --compress
moai-adk backup --restore .moai-backups/latest.tar.gz
```

______________________________________________________________________

## 退出代码

| 代码  | 含义         |
| ----- | ------------ |
| `0`   | 成功         |
| `1`   | 一般错误     |
| `2`   | 用法错误     |
| `127` | 命令不存在   |

______________________________________________________________________

## 环境变量

```bash
MOAI_HOME              MoAI-ADK 安装路径
MOAI_DEBUG             启用调试模式 (1)
MOAI_NO_COLOR          禁用颜色输出 (1)
MOAI_CONFIG_PATH       .moai/config.json 路径
```

______________________________________________________________________

## 故障排除

### "Permission denied"

```bash
# 检查权限
ls -la .moai/
chmod -R u+w .moai/
```

### "Template conflict"

```bash
# 备份后强制更新
moai-adk backup
moai-adk update --force
```

### "Python version mismatch"

```bash
# 确认 Python 3.13+
python3 --version
uv python install 3.13
```

______________________________________________________________________

**下一步**: [Alfred 子命令指南](subcommands.md) 或 [CLI 参考](index.md)
