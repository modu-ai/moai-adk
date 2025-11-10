---
title: リリースプロセス
description: MoAI-ADKバージョン管理とリリース自動化ガイド
status: stable
---

# リリースプロセス

MoAI-ADKのバージョン管理とリリース手順を説明します。

## バージョン管理戦略

MoAI-ADKは[Semantic Versioning](https://semver.org/)に従います：

```
MAJOR.MINOR.PATCH

例: 0.20.1
    │  │   │
    │  │   └─ PATCH: バグ修正（互換性維持）
    │  └────── MINOR: 機能追加（下位互換性維持）
    └───────── MAJOR: 主要な変更（互換性破壊）
```

## リリースサイクル

### 開発段階（developブランチ）

```
1. 機能ブランチで開発
   feature/SPEC-XXX

2. developにPR作成およびマージ
   レビュー → CI/CDチェック → マージ

3. developブランチに機能を蓄積
   複数の機能とバグ修正を含む
```

### リリース準備（release/ブランチ）

```
1. developからreleaseブランチを作成
   git checkout -b release/v0.20.0

2. バージョン更新
   - src/moai_adk/__init__.py: __version__
   - pyproject.toml: version
   - CHANGELOG.md: リリースノート

3. 最終テストとバグ修正
   releaseブランチでのみ修正

4. mainにPR作成
```

### リリースデプロイ（mainブランチ）

```
1. PR承認およびマージ（main）
   git merge release/v0.20.0

2. タグ作成
   git tag -a v0.20.0 -m "Release v0.20.0"

3. PyPIデプロイ自動化
   GitHub Actionsが自動実行

4. developに逆マージ
   main → develop同期
```

## Alfredを使用したリリース

MoAI-ADKはリリース自動化を提供します：

```bash
# パッチリリース（0.20.0 → 0.20.1）
/alfred:release-new patch

# マイナーリリース（0.20.0 → 0.21.0）
/alfred:release-new minor

# メジャーリリース（0.20.0 → 1.0.0）
/alfred:release-new major

# テストモード（実際のデプロイなし）
/alfred:release-new patch --dry-run

# TestPyPIにデプロイ（テスト）
/alfred:release-new patch --testpypi
```

## CHANGELOGの作成

`CHANGELOG.md` の形式：

```markdown
## [0.20.1] - 2025-11-07

### Added
- 新機能1
- 新機能2

### Fixed
- バグ修正1
- バグ修正2

### Changed
- 変更1
- 変更2

### Deprecated
- 非推奨機能

### Security
- セキュリティ関連の修正
```

## バージョン管理ファイル

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

## リリースチェックリスト

リリース前に必ず確認してください：

- [ ] すべての機能がdevelopブランチにマージされている
- [ ] 完全なテストが通過（pytest 100% ✓）
- [ ] コードリンティングが通過（ruff, black, mypy ✓）
- [ ] CHANGELOG.mdが更新されている
- [ ] バージョン番号の一貫性を確認
  - `__init__.py`の`__version__`
  - `pyproject.toml`の`version`
- [ ] READMEおよびドキュメントが最新化されている
- [ ] リリースノートの作成準備

## 自動化されたリリース（GitHub Actions）

`.github/workflows/release.yml` の例：

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

## デプロイ先

### PyPI（プロダクション）

```bash
# 最新リリースをインストール
pip install moai-adk
```

### TestPyPI（テスト）

```bash
# テストデプロイをインストール
pip install -i https://test.pypi.org/simple/ moai-adk
```

### GitHub Releases

- タグベースの自動リリース生成
- リリースノートを含む
- ダウンロード可能なアーティファクト

## 緊急ホットフィックス

緊急バグ修正が必要な場合：

```bash
# mainからhotfixブランチを作成
git checkout main
git checkout -b hotfix/v0.20.2

# バグ修正とコミット
# ... 修正 ...

# mainとdevelopの両方にPR作成
# main: 緊急デプロイ用
# develop: 統合用
```

## リリース担当者

リリースは次の担当者が実行します：

- **Maintainer**: @goos
- **Co-Maintainer**: Community（オプション）

## 参考資料

- [Semantic Versioning](https://semver.org/)
- [Keep a Changelog](https://keepachangelog.com/)
- [Python Packaging Guide](https://packaging.python.org/)

---

**質問がありますか？** GitHub Issuesで質問するか、議論してください！



