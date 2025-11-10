---
title: コードスタイルガイド
description: MoAI-ADK Python, Markdown, YAMLコードスタイル標準
status: stable
---

# コードスタイルガイド

MoAI-ADKのコードスタイル標準を説明します。すべての貢献者はこのガイドに従う必要があります。

## Pythonコードスタイル

### 標準準拠

- **標準**: PEP 8 + Blackフォーマット
- **リンター**: Ruff + mypy（型チェック）
- **フォーマッター**: Black（自動フォーマット）

### ファイル構造

```python
"""
モジュール説明。

このモジュールは... 詳細説明
"""

# 標準ライブラリ
import os
import sys
from pathlib import Path
from typing import Optional

# サードパーティライブラリ
import pytest
from pydantic import BaseModel

# ローカルライブラリ
from moai_adk.core import Agent
from moai_adk.utils import logger


class MyClass:
    """クラス説明。"""

    def method(self) -> None:
        """メソッド説明。"""
        pass
```

### 命名規則

| 項目 | 規則 | 例 |
|------|------|------|
| **クラス** | PascalCase | `class MyAgent:` |
| **関数/メソッド** | snake_case | `def get_config():` |
| **定数** | UPPER_SNAKE_CASE | `DEFAULT_TIMEOUT = 30` |
| **プライベート** | _leading_underscore | `def _internal_method():` |
| **モジュール** | snake_case | `my_module.py` |

### 型ヒント

```python
from typing import Optional, List, Dict, Union

def process_data(
    items: List[str],
    config: Optional[Dict[str, int]] = None,
) -> bool:
    """
    データ処理関数。

    Args:
        items: 処理する項目リスト
        config: オプション設定辞書

    Returns:
        処理成功の有無

    Raises:
        ValueError: 無効な入力
    """
    if not items:
        raise ValueError("items cannot be empty")
    return True
```

### コメントとドキュストリング

```python
def calculate_score(value: int) -> float:
    """
    スコア計算。

    この関数は入力値に基づいて正規化されたスコアを計算します。
    範囲は0.0から1.0の間です。

    Args:
        value: 計算する入力値（0-100）

    Returns:
        正規化されたスコア（0.0-1.0）

    Examples:
        >>> calculate_score(50)
        0.5
    """
    # 範囲検証
    if not 0 <= value <= 100:
        raise ValueError(f"Value must be 0-100, got {value}")

    # スコア計算
    return value / 100.0
```

### 行の長さとフォーマット

```python
# Blackデフォルト: 88文字
# 長すぎる場合は自動的に折り返される

def long_function_name(
    param1: str,
    param2: int,
    param3: Optional[Dict[str, Any]] = None,
) -> Tuple[str, int]:
    """長い関数定義の例。"""
    pass
```

## Markdownスタイル

### ファイル構造

```markdown
---
title: ページタイトル
description: ページ説明
status: stable
---

# H1タイトル

すべてのMarkdownファイルはこの構造に従います。

## H2セクション

### H3サブセクション

より深いタイトルは避けます（H4+は使用しない）。

### リスト形式

**箇条書きリスト（bullet points）**:
- 最初の項目
- 2番目の項目
- 3番目の項目

**順序リスト（numbered）**:
1. 最初のステップ
2. 2番目のステップ
3. 3番目のステップ

### 強調

- **太字**（重要な強調）
- *斜体*（用語の強調）
- ` `（インラインコード）
```

### コードブロック

````markdown
```python
# Pythonコード
def hello():
    print("Hello, World!")
```

```bash
# Bashコマンド
uv run pytest
```

```yaml
# YAML設定
key: value
nested:
  item: value
```
````

### テーブル

```markdown
| ヘッダー1 | ヘッダー2 | ヘッダー3 |
|--------|--------|--------|
| 内容A | 内容B | 内容C |
| 内容D | 内容E | 内容F |
```

## YAMLスタイル

### 設定ファイル

```yaml
# コメントは#の後にスペースを1つ
key: value

# ネスト構造は2スペースインデント
parent:
  child: value
  list_item:
    - item1
    - item2

# 複雑な値は複数行で表現
description: |
  複数行
  テキストは
  パイプで表現します。
```

## 自動化されたスタイルチェック

### Ruff（リンティング）

```bash
# スタイルチェック
uv run ruff check src/

# 自動修正
uv run ruff check --fix src/
```

### Black（フォーマット）

```bash
# フォーマット確認
uv run black --check src/

# 自動フォーマット
uv run black src/
```

### mypy（型チェック）

```bash
# 型チェック
uv run mypy src/moai_adk
```

### 統合チェック

```bash
# すべてのチェックを実行
uv run ruff check src/
uv run black --check src/
uv run mypy src/moai_adk
```

## Pre-commit設定

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

## チェックリスト

PR提出前に確認：

- [ ] PythonコードがBlackでフォーマットされている
- [ ] Ruffリンティングが通過（エラーなし）
- [ ] mypy型チェックが通過
- [ ] Markdownファイルが正しい構造である
- [ ] コードにコメントとドキュストリングが追加されている
- [ ] テストコードが含まれている（テストカバレッジ87%+）

## 参考資料

- [PEP 8](https://www.python.org/dev/peps/pep-0008/)
- [Google Python Style Guide](https://google.github.io/styleguide/pyguide.html)
- [Black Code Style](https://black.readthedocs.io/)
- [CommonMark Spec](https://spec.commonmark.org/)

---

**質問がありますか？** GitHub Discussionsでスタイルに関する質問をしてください！



