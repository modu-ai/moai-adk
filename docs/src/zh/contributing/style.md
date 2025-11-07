---
title: 代码风格指南
description: MoAI-ADK Python, Markdown, YAML 代码风格标准
status: stable
---

# 代码风格指南

说明MoAI-ADK的代码风格标准。所有贡献者都必须遵循本指南。

## 🐍 Python代码风格

### 标准遵守

- **标准**: PEP 8 + Black格式化
- **代码检查器**: Ruff + mypy (类型检查)
- **格式化器**: Black (自动格式化)

### 文件结构

```python
"""
模块说明。

此模块... 详细说明
"""

# 标准库
import os
import sys
from pathlib import Path
from typing import Optional

# 第三方库
import pytest
from pydantic import BaseModel

# 本地库
from moai_adk.core import Agent
from moai_adk.utils import logger


class MyClass:
    """类说明。"""

    def method(self) -> None:
        """方法说明。"""
        pass
```

### 命名规则

| 项目 | 规则 | 示例 |
|------|------|------|
| **类** | PascalCase | `class MyAgent:` |
| **函数/方法** | snake_case | `def get_config():` |
| **常量** | UPPER_SNAKE_CASE | `DEFAULT_TIMEOUT = 30` |
| **私有** | _leading_underscore | `def _internal_method():` |
| **模块** | snake_case | `my_module.py` |

### 类型提示

```python
from typing import Optional, List, Dict, Union

def process_data(
    items: List[str],
    config: Optional[Dict[str, int]] = None,
) -> bool:
    """
    数据处理函数。

    Args:
        items: 要处理的项目列表
        config: 可选配置字典

    Returns:
        处理成功与否

    Raises:
        ValueError: 输入无效
    """
    if not items:
        raise ValueError("items cannot be empty")
    return True
```

### 注释与文档字符串

```python
def calculate_score(value: int) -> float:
    """
    计算分数。

    此函数根据输入值计算标准化分数。
    范围在0.0到1.0之间。

    Args:
        value: 要计算的输入值 (0-100)

    Returns:
        标准化分数 (0.0-1.0)

    Examples:
        >>> calculate_score(50)
        0.5
    """
    # 范围验证
    if not 0 <= value <= 100:
        raise ValueError(f"Value must be 0-100, got {value}")

    # 计算分数
    return value / 100.0
```

### 行长度与格式化

```python
# Black默认值: 88字符
# 长度过长时自动换行

def long_function_name(
    param1: str,
    param2: int,
    param3: Optional[Dict[str, Any]] = None,
) -> Tuple[str, int]:
    """长函数定义示例。"""
    pass
```

## 📝 Markdown风格

### 文件结构

```markdown
---
title: 页面标题
description: 页面描述
status: stable
---

# H1标题

所有Markdown文件遵循此结构。

## H2部分

### H3子部分

避免使用更深的标题 (不使用H4+)。

### 列表格式

**无序列表 (bullet points)**:
- 第一项
- 第二项
- 第三项

**有序列表 (numbered)**:
1. 第一步
2. 第二步
3. 第三步

### 强调

- **粗体** (重要强调)
- *斜体* (术语强调)
- ` ` (内联代码)
```

### 代码块

````markdown
```python
# Python代码
def hello():
    print("Hello, World!")
```

```bash
# Bash命令
uv run pytest
```

```yaml
# YAML配置
key: value
nested:
  item: value
```
````

### 表格

```markdown
| 标题1 | 标题2 | 标题3 |
|--------|--------|--------|
| 内容A | 内容B | 内容C |
| 内容D | 内容E | 内容F |
```

## 🔧 YAML风格

### 配置文件

```yaml
# 注释在#后面加一个空格
key: value

# 嵌套结构使用2个空格缩进
parent:
  child: value
  list_item:
    - item1
    - item2

# 复杂值用多行表示
description: |
  多行
  文本
  使用管道符表示。
```

## 🎯 自动化风格检查

### Ruff (代码检查)

```bash
# 风格检查
uv run ruff check src/

# 自动修复
uv run ruff check --fix src/
```

### Black (格式化)

```bash
# 检查格式化
uv run black --check src/

# 自动格式化
uv run black src/
```

### mypy (类型检查)

```bash
# 类型检查
uv run mypy src/moai_adk
```

### 集成检查

```bash
# 运行所有检查
uv run ruff check src/
uv run black --check src/
uv run mypy src/moai_adk
```

## 📋 Pre-commit设置

`.pre-commit-config.yaml`:

```yaml
repos:
  - repo: https://github.com/astral-sh/ruff-pre-commit
    rev: v0.1.0
    hooks:
      - id: ruff
        args: [--fix]

  - repo: https://github.com/psf/black
    rev: 23.10.0
    hooks:
      - id: black

  - repo: https://github.com/pre-commit/mirrors-mypy
    rev: v1.6.0
    hooks:
      - id: mypy
```

## ✅ 检查清单

提交PR前确认:

- [ ] Python代码已用Black格式化
- [ ] 通过Ruff代码检查 (无错误)
- [ ] 通过mypy类型检查
- [ ] Markdown文件结构正确
- [ ] 代码添加了注释和文档字符串
- [ ] 包含测试代码 (测试覆盖率87%+)

## 📚 参考资料

- [PEP 8](https://www.python.org/dev/peps/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [Black Code Style](https://black.readthedocs.io/)
- [CommonMark Spec](https://spec.commonmark.org/)

---

**有问题?** 在GitHub Discussions中提问关于风格的问题！
