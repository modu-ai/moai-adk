---
title: 开发环境设置
description: MoAI-ADK本地开发环境配置与贡献指南
status: stable
---

# 开发环境设置

如何配置用于为MoAI-ADK贡献的本地开发环境。

## 📋 前置条件

- Python 3.13+
- Git
- UV (Python包管理器)
- Docker (可选)

## 🚀 配置开发环境

### 第1步: 克隆仓库

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### 第2步: 安装开发依赖

```bash
# 使用UV安装 (推荐)
uv sync --all-extras

# 或使用pip
pip install -e ".[dev,test,docs]"
```

### 第3步: 设置预提交钩子

```bash
# 安装Pre-commit钩子
uv run pre-commit install

# 对所有文件运行预检查
uv run pre-commit run --all-files
```

## 🧪 运行测试

### 完整测试套件

```bash
# 运行所有测试
uv run pytest

# 包含覆盖率报告
uv run pytest --cov=src/moai_adk --cov-report=html
```

### 运行特定测试

```bash
# 测试特定文件
uv run pytest tests/test_core.py

# 测试特定函数
uv run pytest tests/test_core.py::test_function_name

# 基于标记运行
uv run pytest -m integration
```

## 📝 代码风格检查

### 代码检查

```bash
# 使用Ruff进行代码检查
uv run ruff check src/ tests/

# 使用Black进行格式化
uv run black src/ tests/

# 使用mypy进行类型检查
uv run mypy src/moai_adk
```

### 自动修复

```bash
# Ruff自动修复
uv run ruff check --fix src/ tests/

# Black自动格式化
uv run black src/ tests/
```

## 📚 构建文档

### 本地文档服务器

```bash
cd docs

# 启动开发服务器
uv run mkdocs serve

# 在浏览器中访问 http://localhost:8000
```

### 生产环境构建

```bash
# 生成静态网站
uv run mkdocs build

# 输出: site/ 目录
```

## 🔧 开发工作流

### 创建功能分支

```bash
# 同步最新的develop分支
git checkout develop
git pull origin develop

# 创建功能分支
git checkout -b feature/SPEC-XXX

# 或使用Alfred
/alfred:1-plan "功能标题"
```

### 本地开发与测试

```bash
# 编写代码
# ... 修改工作 ...

# 运行测试
uv run pytest

# 代码风格检查
uv run ruff check --fix src/
uv run black src/

# 类型检查
uv run mypy src/moai_adk
```

### 提交与推送

```bash
# 添加更改
git add .

# 使用Alfred提交 (推荐)
/alfred:2-run SPEC-XXX

# 或手动提交
git commit -m "feat: 功能描述"
git push origin feature/SPEC-XXX
```

## 🔄 Pull Request流程

1. **创建PR**: 从功能分支向develop创建PR
2. **自动检查**: GitHub Actions自动运行测试和代码检查
3. **代码审查**: 等待维护者审查
4. **修改请求**: 必要时反映反馈
5. **合并**: 批准后合并到develop分支

## 🐛 调试

### 设置日志级别

```bash
# 启用调试模式
export MOAI_DEBUG=true
uv run moai-adk init my-project
```

### VS Code调试

`.vscode/launch.json` 示例:

```json
{
  "version": "0.2.0",
  "configurations": [
    {
      "name": "Python: Current File",
      "type": "python",
      "request": "launch",
      "program": "${file}",
      "console": "integratedTerminal"
    }
  ]
}
```

## 📚 参考文档

- [代码风格指南](style.md)
- [发布流程](releases.md)
- [贡献者行为准则](index.md)

## ❓ 问题排查

### 依赖项错误

```bash
# 清除缓存并重新安装
uv cache clean
uv sync --all-extras
```

### 测试失败

```bash
# 以详细输出运行测试
uv run pytest -vv

# 仅运行特定测试
uv run pytest tests/test_xxx.py::test_name -vv
```

### 文档构建错误

```bash
# 清除缓存
rm -rf docs/site docs/.cache

# 重新构建
cd docs
uv run mkdocs build --strict
```

---

**有问题?** 在GitHub Issues中提问或参与Discussions！
