---
title: 開発環境設定
description: MoAI-ADKローカル開発環境の構成と貢献ガイド
status: stable
---

# 開発環境設定

MoAI-ADKに貢献するためのローカル開発環境を構成する方法を説明します。

## 前提条件

- Python 3.13+
- Git
- UV (Pythonパッケージマネージャー)
- Docker (オプション)

## 開発環境の構成

### ステップ1: リポジトリのクローン

```bash
git clone https://github.com/modu-ai/moai-adk.git
cd moai-adk
```

### ステップ2: 開発依存関係のインストール

```bash
# UVを使用したインストール（推奨）
uv sync --all-extras

# またはpipを使用
pip install -e ".[dev,test,docs]"
```

### ステップ3: プリコミットフックの設定

```bash
# Pre-commitフックのインストール
uv run pre-commit install

# すべてのファイルに対して事前チェックを実行
uv run pre-commit run --all-files
```

## テストの実行

### 完全なテストスイート

```bash
# すべてのテストを実行
uv run pytest

# カバレッジレポートを含む
uv run pytest --cov=src/moai_adk --cov-report=html
```

### 特定のテストの実行

```bash
# 特定のファイルをテスト
uv run pytest tests/test_core.py

# 特定の関数をテスト
uv run pytest tests/test_core.py::test_function_name

# マーカーに基づいて実行
uv run pytest -m integration
```

## コードスタイルチェック

### リンティング

```bash
# Ruffでリンティング
uv run ruff check src/ tests/

# Blackでフォーマット
uv run black src/ tests/

# mypyで型チェック
uv run mypy src/moai_adk
```

### 自動修正

```bash
# Ruff自動修正
uv run ruff check --fix src/ tests/

# Black自動フォーマット
uv run black src/ tests/
```

## ドキュメントのビルド

### ローカルドキュメントサーバー

```bash
cd docs

# 開発サーバーを起動
uv run mkdocs serve

# ブラウザで http://localhost:8000 にアクセス
```

### プロダクションビルド

```bash
# 静的サイトを生成
uv run mkdocs build

# 出力: site/ ディレクトリ
```

## 開発ワークフロー

### 機能ブランチの作成

```bash
# 最新のdevelopブランチを同期
git checkout develop
git pull origin develop

# 機能ブランチを作成
git checkout -b feature/SPEC-XXX

# またはAlfredを使用
/alfred:1-plan "機能タイトル"
```

### ローカル開発とテスト

```bash
# コードを書く
# ... 変更作業 ...

# テストを実行
uv run pytest

# コードスタイルチェック
uv run ruff check --fix src/
uv run black src/

# 型チェック
uv run mypy src/moai_adk
```

### コミットとプッシュ

```bash
# 変更を追加
git add .

# Alfredを使用したコミット（推奨）
/alfred:2-run SPEC-XXX

# または手動コミット
git commit -m "feat: 機能説明"
git push origin feature/SPEC-XXX
```

## Pull Requestプロセス

1. **PR作成**: 機能ブランチからdevelopへのPR作成
2. **自動チェック**: GitHub Actionsが自動テストとリンティングを実行
3. **コードレビュー**: メンテナーのレビューを待つ
4. **マージ**: 承認後、developブランチにマージ

## デバッグ

### ログレベルの設定

```bash
# デバッグモードを有効化
export MOAI_DEBUG=true
uv run moai-adk init my-project
```

### VS Codeデバッグ

`.vscode/launch.json` の例:

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

## 参考ドキュメント

- [コードスタイルガイド](style.md)
- [リリースプロセス](releases.md)
- [貢献者行動規範](index.md)

## トラブルシューティング

### 依存関係エラー

```bash
# キャッシュをクリアして再インストール
uv cache clean
uv sync --all-extras
```

### テスト失敗

```bash
# 詳細な出力でテストを実行
uv run pytest -vv

# 特定のテストのみを実行
uv run pytest tests/test_xxx.py::test_name -vv
```

### ドキュメントビルドエラー

```bash
# キャッシュをクリア
rm -rf docs/site docs/.cache

# 再ビルド
cd docs
uv run mkdocs build --strict
```

---

**質問がありますか？** GitHub Issuesで質問するか、Discussionsに参加してください！



