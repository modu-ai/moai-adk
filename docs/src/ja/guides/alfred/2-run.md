______________________________________________________________________

## title: /alfred:2-run コマンド description: TDD実装と品質保証のための完全ガイド lang: ja

# /alfred:2-run - TDD実装コマンド

`/alfred:2-run`はMoAI-ADKの実装段階コマンドで、Test-Driven Development (TDD)サイクルを通じてコードを実装し、品質を保証します。

## 概要

**目的**: TDD RED→GREEN→REFACTORサイクルによる実装 **実行時間**: 約5-10分 **主要成果**: 100%テストカバレッジのコード、品質検証通過

## 基本使用法

```bash
/alfred:2-run SPEC-ID
```

### 例

```bash
# SPEC-ID指定で実装
/alfred:2-run HELLO-001
/alfred:2-run AUTH-001
/alfred:2-run USER-002
```

## TDDサイクル

### 🔴 RED Phase - 失敗するテスト作成

**目標**: コード作成前にテストを先に作成

**実行内容**:
1. SPECのすべての要件をテストケースに変換
2. エッジケースと例外状況を含む
3. テストが失敗することを確認

**例**:
```python
# tests/test_hello.py
# @TEST:EX-HELLO-001 | SPEC: SPEC-HELLO-001.md

def test_hello_with_name_should_return_personalized_greeting():
    """nameが提供されたら 'Hello, {name}!'を返すべきである"""
    response = client.get("/hello?name=Alice")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, Alice!"}

def test_hello_without_name_should_return_default_greeting():
    """nameがない場合 'Hello, World!'を返すべきである"""
    response = client.get("/hello")
    assert response.status_code == 200
    assert response.json() == {"message": "Hello, World!"}
```

**Git Commit**:
```bash
git commit -m "🔴 test(HELLO-001): add failing hello API tests"
```

### 🟢 GREEN Phase - 最小実装

**目標**: テストを通過させる最小限のコード作成

**実行内容**:
1. テストを通過させる最小限の実装
2. コード品質より機能動作を優先
3. すべてのテストが通過することを確認

**例**:
```python
# src/hello/api.py
# @CODE:EX-HELLO-001:API | SPEC: SPEC-HELLO-001.md | TEST: tests/test_hello.py

from fastapi import FastAPI

app = FastAPI()

@app.get("/hello")
def hello(name: str = "World"):
    """@CODE:EX-HELLO-001:API - Helloエンドポイント"""
    return {"message": f"Hello, {name}!"}
```

**Git Commit**:
```bash
git commit -m "🟢 feat(HELLO-001): implement hello API"
```

### ♻️ REFACTOR Phase - コード品質改善

**目標**: テストを維持しながらコード品質向上

**実行内容**:
1. TRUST 5原則適用
2. コード構造改善
3. パフォーマンス最適化
4. ドキュメント追加

**例**:
```python
# src/hello/models.py (新規)
# @CODE:EX-HELLO-001:MODEL | SPEC: SPEC-HELLO-001.md

from pydantic import BaseModel, Field, validator

class HelloRequest(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - リクエスト検証モデル"""
    name: str = Field(
        default="World", 
        max_length=50,
        description="挨拶する名前"
    )
    
    @validator('name')
    def validate_name(cls, v):
        if not v.strip():
            raise ValueError('名前は空にできません')
        return v.strip()

class HelloResponse(BaseModel):
    """@CODE:EX-HELLO-001:MODEL - レスポンスモデル"""
    message: str = Field(description="挨拶メッセージ")
```

**Git Commit**:
```bash
git commit -m "♻️ refactor(HELLO-001): add models and improve validation"
```

## 実装計画段階

### implementation-plannerの役割

実装前に以下を分析:

1. **アーキテクチャ設計**
   - フォルダ構造
   - ファイル配置
   - 依存関係

2. **技術スタック決定**
   - フレームワーク選択
   - ライブラリ推薦
   - ツール設定

3. **TAG戦略**
   - TAG命名規則
   - TAGチェーン設計
   - 追跡性保証

4. **専門家活性化**
   - 必要な専門家判断
   - コンサルテーション実行
   - 推薦事項統合

## 品質検証

### TRUST 5原則

#### 1. Test First ✅
- **テストカバレッジ**: 85%以上必須
- **テスト種類**: ユニット、統合、E2E
- **エッジケース**: すべての例外状況

#### 2. Readable ✅
- **関数長**: 50行以下
- **変数命名**: 明確で意味のある名前
- **ドキュメント**: すべての公開関数に

#### 3. Unified ✅
- **アーキテクチャ**: 一貫したパターン
- **API設計**: RESTful規約準拠
- **エラーハンドリング**: 統一された形式

#### 4. Secured ✅
- **入力検証**: すべてのユーザー入力
- **認証/認可**: 適切な保護
- **機密データ**: 安全な処理

#### 5. Trackable ✅
- **TAG使用**: すべてのコードに@TAG
- **Git履歴**: 明確なコミットメッセージ
- **ドキュメント**: コードとドキュメントの連携

### 自動品質検査

```bash
# テスト実行
pytest tests/ -v --cov=src

# コードスタイル検査
ruff check src/

# タイプチェック
mypy src/

# セキュリティスキャン
bandit -r src/
```

## 生成される成果物

### 1. テストファイル
**場所**: `tests/test_{module}.py`
**TAG**: `@TEST:EX-{ID}`
**内容**: すべての要件に対するテストケース

### 2. 実装コード
**場所**: `src/{module}/`
**TAG**: `@CODE:EX-{ID}:{SUBTYPE}`
**内容**: プロダクション品質コード

### 3. Git履歴
**コミット**: RED → GREEN → REFACTOR
**メッセージ**: 明確なタイプとSPEC-ID

## 高度な機能

### 並列テスト実行

```bash
# 複数コアで並列実行
pytest -n auto

# 特定テストのみ
pytest tests/test_auth.py -k "login"
```

### カバレッジレポート

```bash
# HTML レポート生成
pytest --cov=src --cov-report=html

# ブラウザで確認
open htmlcov/index.html
```

### 段階的実装

```bash
# 特定機能だけ実装
/alfred:2-run AUTH-001 --feature login

# 次の機能
/alfred:2-run AUTH-001 --feature logout
```

## トラブルシューティング

### テストが失敗する場合

```bash
# 詳細出力で実行
pytest -vv

# 失敗したテストのみ再実行
pytest --lf

# デバッグモード
pytest --pdb
```

### 依存関係問題

```bash
# 依存関係再インストール
uv sync --force

# 特定パッケージ追加
uv add fastapi pytest
```

### タイムアウトエラー

```bash
# タイムアウト延長
pytest --timeout=300

# 特定テストスキップ
pytest -m "not slow"
```

## ベストプラクティス

### 1. 小さなステップで進む
- 一度に一つの機能
- RED-GREEN-REFACTORサイクル厳守
- 頻繁なコミット

### 2. テストの品質
- 明確なテスト名
- 独立したテストケース
- モックの適切な使用

### 3. コードレビュー
- 自己レビュー実施
- ペアプログラミング推奨
- CI/CD統合

## チーム協業

### コードレビュープロセス

1. **実装完了**: `/alfred:2-run`完了
2. **自己レビュー**: 品質基準確認
3. **PR作成**: Draft PRオープン
4. **チームレビュー**: フィードバック収集
5. **修正**: レビューコメント反映
6. **承認**: マージ承認

### 品質ゲート

```yaml
# CI/CD品質ゲート
quality_gates:
  test_coverage: 85%
  code_quality: A
  security_scan: pass
  lint_check: pass
  type_check: pass
```

______________________________________________________________________

**📚 次のステップ**:

- [/alfred:3-sync](3-sync.md)でドキュメント同期
- [TDDガイド](../tdd/index.md)でテスト技術深化
- [品質保証](../quality/index.md)でコード品質向上
