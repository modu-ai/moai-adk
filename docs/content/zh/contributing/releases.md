---
title: 发布流程
description: MoAI-ADK版本管理与发布自动化指南
status: stable
---

# 发布流程

说明MoAI-ADK的版本管理和发布程序。

## 版本管理策略

MoAI-ADK遵循[语义化版本](https://semver.org/)：

```
MAJOR.MINOR.PATCH

例: 0.20.1
    │  │   │
    │  │   └─ PATCH: 错误修复（保持兼容性）
    │  └────── MINOR: 功能添加（保持向后兼容性）
    └───────── MAJOR: 重大更改（破坏兼容性）
```

## 发布周期

### 开发阶段（develop分支）

```
1. 在功能分支中开发
   feature/SPEC-XXX

2. 创建PR并合并到develop
   审查 → CI/CD检查 → 合并

3. 在develop分支中积累功能
   包含多个功能和错误修复
```

### 发布准备（release/分支）

```
1. 从develop创建release分支
   git checkout -b release/v0.20.0

2. 更新版本
   - src/moai_adk/__init__.py: __version__
   - pyproject.toml: version
   - CHANGELOG.md: 发布说明

3. 最终测试和错误修复
   仅在release分支中修复

4. 创建PR到main
```

### 发布部署（main分支）

```
1. PR批准并合并（main）
   git merge release/v0.20.0

2. 创建标签
   git tag -a v0.20.0 -m "Release v0.20.0"

3. PyPI部署自动化
   GitHub Actions自动执行

4. 反向合并到develop
   main → develop同步
```

## 使用Alfred进行发布

MoAI-ADK提供发布自动化：

```bash
# 补丁发布（0.20.0 → 0.20.1）
/alfred:release-new patch

# 次要发布（0.20.0 → 0.21.0）
/alfred:release-new minor

# 主要发布（0.20.0 → 1.0.0）
/alfred:release-new major

# 测试模式（不实际部署）
/alfred:release-new patch --dry-run

# 部署到TestPyPI（测试）
/alfred:release-new patch --testpypi
```

## CHANGELOG编写

`CHANGELOG.md` 格式：

```markdown
## [0.20.1] - 2025-11-07

### Added
- 新功能1
- 新功能2

### Fixed
- 错误修复1
- 错误修复2

### Changed
- 更改1
- 更改2

### Deprecated
- 已弃用功能

### Security
- 安全相关修复
```

## 版本管理文件

### src/moai_adk/__init__.py

```python
"""
MoAI-ADK: Agentic Development Kit
"""

__version__ = "0.20.1"
__author__ = "GoosLab"
__license__ = "MIT"
```

### pyproject.toml

```toml
[project]
name = "moai-adk"
version = "0.20.1"
description = "MoAI-Agentic Development Kit"
```

## 发布检查清单

发布前必须确认：

- [ ] 所有功能已合并到develop分支
- [ ] 所有测试通过（pytest 100% ✓）
- [ ] 代码检查通过（ruff, black, mypy ✓）
- [ ] CHANGELOG.md已更新
- [ ] 版本号一致性确认
  - `__init__.py`的`__version__`
  - `pyproject.toml`的`version`
- [ ] README和文档已更新
- [ ] 发布说明准备就绪

## 自动化发布（GitHub Actions）

`.github/workflows/release.yml` 示例：

```yaml
name: Release

on:
  push:
    tags:
      - 'v*'

jobs:
  release:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Build
        run: uv build

      - name: Publish to PyPI
        run: uv publish
        env:
          UV_PUBLISH_TOKEN: ${{ secrets.PYPI_TOKEN }}
```

## 部署目标

### PyPI（生产环境）

```bash
# 安装最新发布
pip install moai-adk
```

### TestPyPI（测试）

```bash
# 安装测试部署
pip install -i https://test.pypi.org/simple/ moai-adk
```

### GitHub Releases

- 基于标签的自动发布生成
- 包含发布说明
- 可下载的工件

## 紧急热修复

需要紧急错误修复时：

```bash
# 从main创建hotfix分支
git checkout main
git checkout -b hotfix/v0.20.2

# 错误修复和提交
# ... 修复 ...

# 创建PR到main和develop
# main: 紧急部署用
# develop: 集成用
```

## 发布负责人

发布由以下负责人执行：

- **Maintainer**: @goos
- **Co-Maintainer**: Community（可选）

## 参考资料

- [语义化版本](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Python打包指南](https://packaging.python.org/)

---

**有问题?** 在GitHub Issues中提问或讨论！



