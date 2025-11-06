---
title: /alfred:2-run コマンド
description: TDD実装と品質保証のための完全ガイド
lang: ja
---

# /alfred:2-run - TDD実装コマンド

`/alfred:2-run`はMoAI-ADKの実装段階コマンドで、テスト駆動開発(TDD)サイクルを自動実行し、高品質なコードを作成します。

## 概要

**目的**: TDDサイクル実行と品質保証
**実行時間**: 約5分
**主要成果**: テストコード、実装コード、品質レポート

## 基本使用法

```bash
/alfred:2-run SPEC-ID
```

### 例

```bash
# Hello API実装
/alfred:2-run HELLO-001

# ユーザー認証実装
/alfred:2-run AUTH-001

# TODO機能実装
/alfred:2-run TODO-001
```

## TDDサイクルの自動実行

### フェーズ1: 実装計画立案

#### implementation-plannerが自動実行

1. **SPEC分析**: 要件と制約事項を分析
2. **技術選定**: 適切なライブラリとフレームワークを選択
3. **アーキテクチャ設計**: フォルダ構造とコンポーネント設計
4. **TAG設計**: 追跡のためのTAG ID割り当て

#### 計画出力例

```
<span class="material-icons">construction</span> implementation-plannerの分析結果:

## 技術スタック選定
- **フレームワーク**: FastAPI (高性能、自動ドキュメンテーション)
- **検証**: Pydantic (タイプ安全、入力検証)
- **テスト**: pytest (豊富な機能、広く使われている)
- **データベース**: SQLite (開発段階、後でPostgreSQLに移行可能)

## フォルダ構造
```
src/hello/
├── __init__.py
├── api.py          # @CODE:EX-HELLO-001:API
├── models.py       # @CODE:EX-HELLO-001:MODEL
└── services.py     # @CODE:EX-HELLO-001:SERVICE

tests/hello/
├── __init__.py
├── test_api.py     # @TEST:EX-HELLO-001
└── test_models.py  # @TEST:EX-HELLO-002
```

## TAG設計
- @CODE:EX-HELLO-001:API - APIエンドポイント実装
- @CODE:EX-HELLO-001:MODEL - データモデル定義
- @CODE:EX-HELLO-001:SERVICE - ビジネスロジック
- @TEST:EX-HELLO-001 - APIテスト
- @TEST:EX-HELLO-002 - モデルテスト
```

### フェーズ2: TDDサイクル実行

#### 🔴 RED: 失敗するテスト作成

tdd-implementerがSPECからテストケースを自動生成：

```python
# tests/test_hello.py
# @TEST:EX-HELLO-001 | SPEC: SPEC-HELLO-001.md

import pytest
from fastapi.testclient import TestClient
from src.hello.api import app

client = TestClient(app)

def test_hello_with_name_should_return_personalized_greeting():
    """WHEN クエリパラメータnameが提供されたら、"Hello, {name}!"を返すべきである"""
    response = client.get("/hello?name=田中")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, 田中!"}

def test_hello_without_name_should_return_default_greeting():
    """WHEN nameがない場合、"Hello, World!"を返すべきである"""
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}

def test_hello_with_long_name_should_return_400():
    """nameが50文字を超える場合、400エラーを返すべきである"""
    long_name = "a" * 51
    response = client.get(f"/hello?name={long_name}")
    assert response.status_code == 400
    assert "too long" in response.json()["detail"].lower()

def test_hello_with_invalid_chars_should_return_400():
    """無効な文字が含まれる場合、400エラーを返すべきである"""
    response = client.get("/hello?name=<script>")
    assert response.status_code == 400
```

**実行結果**: <span class="material-icons">cancel</span> FAILED (予期通り - 実装がまだない)

**Gitコミット**:
```bash
git add tests/test_hello.py
git commit -m "🔴 test(HELLO-001): add failing hello API tests"
```

#### 🟢 GREEN: 最小実装

テストを通過させる最小のコードを実装：

```python
# src/hello/api.py
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException
from pydantic import BaseModel
import re

app = FastAPI()

class HelloResponse(BaseModel):
    message: str

@app.get("/hello", response_model=HelloResponse)
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Helloエンドポイント"""

    # 制約検証
    if len(name) > 50:
        raise HTTPException(status_code=400, detail="Name too long (max 50 chars)")

    # 無効な文字検証（基本的なXSS防止）
    if re.search(r'[<>"\']', name):
        raise HTTPException(status_code=400, detail="Invalid characters in name")

    return {"message": f"Hello, {name}!"}
```

**実行結果**: <span class="material-icons">check_circle</span> PASSED (すべてのテスト通過)

**Gitコミット**:
```bash
git add src/hello/api.py
git commit -m "🟢 feat(HELLO-001): implement hello API with validation"
```

#### <span class="material-icons">recycling</span> REFACTOR: コード改善

TRUST 5原則を適用してコードを改善：

```python
# src/hello/models.py
# @CODE:EX-HELLO-001:MODEL | SPEC: SPEC-HELLO-001.md

from pydantic import BaseModel, Field, validator
import re

class HelloRequest(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - Helloリクエストモデル"""
    name: str = Field(
        default="World",
        min_length=1,
        max_length=50,
        description="挨拶する名前"
    )

    @validator('name')
    def validate_name(cls, v):
        """名前の妥当性検証"""
        if not v.strip():
            raise ValueError('Name cannot be empty')

        # XSS防止のため無効な文字を検証
        if re.search(r'[<>"\']', v):
            raise ValueError('Name contains invalid characters')

        return v.strip()

class HelloResponse(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - Helloレスポンスモデル"""
    message: str = Field(..., description="挨拶メッセージ")
```

```python
# src/hello/services.py
# @CODE:EX-HELLO-001:SERVICE | SPEC: SPEC-HELLO-001.md

from .models import HelloRequest, HelloResponse

class HelloService:
    """@CODE:EX-HELLO-001:SERVICE - Helloビジネスロジック"""

    @staticmethod
    def create_greeting(name: str = "World") -> HelloResponse:
        """挨拶メッセージを作成する"""
        request = HelloRequest(name=name)
        message = f"Hello, {request.name}!"
        return HelloResponse(message=message)
```

```python
# src/hello/api.py (改善版)
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI, HTTPException, Query
from .services import HelloService
from .models import HelloResponse

app = FastAPI(
    title="Hello API",
    description="簡単な挨拶API",
    version="1.0.0"
)

@app.get("/hello", response_model=HelloResponse)
def hello(
    name: str = Query(
        default="World",
        min_length=1,
        max_length=50,
        description="挨拶する名前"
    )
):
    """@CODE:EX-HELLO-001:API - Helloエンドポイント（改善版）"""
    try:
        return HelloService.create_greeting(name)
    except ValueError as e:
        raise HTTPException(status_code=400, detail=str(e))
```

**実行結果**: <span class="material-icons">check_circle</span> PASSED (すべてのテスト依然として通過)

**Gitコミット**:
```bash
git add src/hello/models.py src/hello/services.py src/hello/api.py
git commit -m "<span class="material-icons">recycling</span> refactor(HELLO-001): improve code structure with TRUST principles"
```

### フェーズ3: 品質検証

#### quality-gateが自動検証実行

```
<span class="material-icons">shield</span> quality-gateの検証結果:

## TRUST 5原則検証
<span class="material-icons">check_circle</span> Test First: カバレッジ 95% (最小要件: 85%)
<span class="material-icons">check_circle</span> Readable: すべての関数 < 30行、変数名が意図を表す
<span class="material-icons">check_circle</span> Unified: SPECベースアーキテクチャ維持、一貫したパターン
<span class="material-icons">check_circle</span> Secured: 入力検証完了、XSS防止、エラーメッセージ安全
<span class="material-icons">check_circle</span> Trackable: すべてのファイルに@TAG割り当て完了

## コード品質メトリクス
- サイクロマティック複雑度: 3 (優秀)
- 重複コード: 0% (完璧)
- テストカバレッジ: 95% (優秀)
- タイプ安全性: 100% (Pydantic使用)

## パフォーマンス予測
- 予想レスポンスタイム: < 10ms
- メモリ使用量: < 50MB
- 同時実行能力: 1000+ リクエスト/秒
```

## 高度な機能

### 並列テスト実行

```bash
# 複数SPECを同時に実行
/alfred:2-run AUTH-001 USER-001 TODO-001

# 並列実行ログ
🔄 並列TDD実行開始:
  ├─ AUTH-001: tdd-implementer活性化 (進行中: 0%)
  ├─ USER-001: tdd-implementer活性化 (進行中: 0%)
  └─ TODO-001: tdd-implementer活性化 (進行中: 0%)
```

### カスタムテスト戦略

```bash
# 特定テストタイプに焦点
/alfred:2-run HELLO-001 --focus=integration

# パフォーマステスト追加
/alfred:2-run HELLO-001 --add-performance-tests

# セキュリティテスト追加
/alfred:2-run AUTH-001 --add-security-tests
```

### データベース統合

```bash
# データベース必須機能
/alfred:2-run USER-001 --database=postgresql

# 生成されるデータベーステスト
def test_user_crud_with_database():
    """データベースでのCRUD操作テスト"""
    # テスト用データベース設定
    # CRUD操作テスト
    # データ整合性検証
```

## 専門家との連携

### 自動専門家活性化

特定の状況で専門家を自動的に活性化：

| 状況 | 活性化される専門家 | 提供内容 |
|------|------------------|----------|
| データベース関連機能 | database-expert | スキーマ設計、クエリ最適化 |
| 認証・認可機能 | security-expert | セキュリティ実装、脆弱性分析 |
| APIエンドポイント | backend-expert | API設計、ドキュメンテーション |
| パフォーマンス要件 | devops-expert | 性能最適化、スケーリング |

### 専門家アドバイス統合

```
<span class="material-icons">settings</span> backend-expertの実装アドバイス:
- APIバージョニングを追加することを推奨
- レート制限ミドルウェアを検討
- OpenAPIドキュメンテーションを自動生成

<span class="material-icons">storage</span> database-expertのアドバイス:
- ユニーク制約とインデックスを追加
- トランザクション管理を検討
- データベースマイグレーション戦略が必要

<span class="material-icons">lock</span> security-expertのアドバイス:
- ログ記録と監査を追加
- レート制限で悪用防止
- エラーメッセージから情報漏洩防止
```

## テスト戦略

### テストピラミッド

```
    /\
   /  \     E2Eテスト (10%)
  /____\
 /      \   統合テスト (20%)
/________\  単体テスト (70%)
```

### テストタイプ

1. **単体テスト**: 個別関数/メソッドテスト
2. **統合テスト**: コンポーネント間連携テスト
3. **E2Eテスト**: 完全なユーザーシナリオテスト

### カバレッジ要件

- **最小要件**: 85%
- **推奨**: 90%以上
- **目標**: 95%以上

## 品質メトリクス

### TRUST 5原則詳細

#### 🧪 Test First

```bash
# カバレッジ確認
pytest --cov=src tests/
# 期待: coverage >= 85%

# カバレッジレポート
coverage html
# HTMLレポート生成: htmlcov/index.html
```

#### <span class="material-icons">auto_stories</span> Readable

```python
# 良い例: 明確な関数名と変数名
def create_user_with_email_validation(user_data: dict) -> User:
    """メール検証付きユーザー作成"""
    validated_email = validate_email_format(user_data["email"])
    return User(email=validated_email, **user_data)

# 悪い例: 曖昧な名前
def process(data):
    # 何を処理しているか不明
    return something
```

#### <span class="material-icons">target</span> Unified

- **アーキテクチャ一貫性**: すべてのレイヤーで同じパターン使用
- **命名規則**: ファイル、関数、変数名の一貫性
- **エラー処理**: 全体的なエラー処理戦略

#### <span class="material-icons">lock</span> Secured

```python
# 入力検証例
from pydantic import BaseModel, validator

class UserInput(BaseModel):
    email: str
    password: str

    @validator('email')
    def validate_email(cls, v):
        email_regex = r'^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$'
        if not re.match(email_regex, v):
            raise ValueError('Invalid email format')
        return v.lower()

    @validator('password')
    def validate_password(cls, v):
        if len(v) < 8:
            raise ValueError('Password must be at least 8 characters')
        return v
```

#### <span class="material-icons">link</span> Trackable

```python
# すべてのファイルにTAG付与
# @CODE:EX-USER-001:MODEL | SPEC: SPEC-USER-001.md | TEST: tests/test_user_models.py

class User(BaseModel):
    """@CODE:EX-USER-001:MODEL - ユーザーデータモデル"""
    # 実装内容...
```

## トラブルシューティング

### よくある問題

**テストが失敗し続ける**:
```bash
# テストデバッグ
pytest tests/test_hello.py -v -s

# 特定テストのみ実行
pytest tests/test_hello.py::test_hello_with_name_should_return_personalized_greeting -v
```

**モジュールインポートエラー**:
```bash
# 依存関係インストール
uv add fastapi pytest

# Pythonパス確認
python -c "import sys; print(sys.path)"
```

**カバレッジ不足**:
```bash
# カバレッジレポート確認
coverage report -m

# 未カバレッジライン確認
coverage html
# htmlcov/index.htmlを開く
```

## ベストプラクティス

### 1. 小さなステップ

- **一つのテスト**: 一つの機能のみテスト
- **小さな実装**: テストを通過させる最小コード
- **漸進的改善**: 少しずつコードを改善

### 2. 明確なテスト名

```python
# 良い例: 何を、いつ、どのようにテストするか明確
def test_user_creation_with_valid_email_should_return_user_object()

# 悪い例: 曖昧で何をテストするか不明
def test_user()
```

### 3. テスト独立性

```python
# 各テストが独立している
def test_create_user():
    user = User(name="田中")
    assert user.name == "田中"

def test_update_user():
    user = User(name="田中")
    user.name = "佐藤"
    assert user.name == "佐藤"  # 他のテストに依存しない
```

### 4. 適切なアサーション

```python
# 具体的なアサーション
assert response.status_code == 200
assert response.json()["message"] == "Hello, 田中!"

# 曖昧なアサーション（避ける）
assert response.ok
assert "message" in response.json()
```

## 統合と連携

### /alfred:1-planとの連携

```bash
# SPECから直接実装
/alfred:1-plan "機能説明"
→ SPEC-ID生成
/alfred:2-run SPEC-ID  # 生成されたIDを使用
```

### /alfred:3-syncとの連携

```bash
# 実装完了後同期
/alfred:2-run SPEC-ID
/alfred:3-sync  # 自動ドキュメント生成
```

### CI/CD統合

```yaml
# .github/workflows/test.yml
name: Test
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Set up Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.13'
      - name: Install dependencies
        run: |
          pip install -r requirements.txt
      - name: Run tests
        run: pytest --cov=src tests/
      - name: Check coverage
        run: |
          coverage report --fail-under=85
```

---

**<span class="material-icons">auto_stories</span> 次のステップ**:
- [/alfred:3-sync](3-sync.md)でドキュメント同期
- [TDDガイド](../tdd/index.md)でテスト駆動開発技術
- [品質ガイド](../project/deploy.md)でプロダクション展開