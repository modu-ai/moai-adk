---
title: /alfred:3-sync コマンド
description: ドキュメント同期と状態管理のための完全ガイド
lang: ja
---

# /alfred:3-sync - ドキュメント同期コマンド

`/alfred:3-sync`はMoAI-ADKの同期段階コマンドで、コード、テスト、ドキュメントを最新状態に同期し、プロジェクトの完全性を保証します。

## 概要

**目的**: ドキュメント同期と状態管理
**実行時間**: 約1分
**主要成果**: Living Document、同期レポート、TAG検証

## 基本使用法

```bash
/alfred:3-sync
# または
/alfred:3-sync auto  # 自動モード
```

### 実行タイミング

- **実装完了後**: TDDサイクル完了後必ず実行
- **コミット前**: 変更をリポジトリに反映する前
- **PR作成時**: プルリクエスト作成前
- **リリース準備時**: 本番環境展開前

## 同期プロセス

### フェーズ1: TAGチェーン検証

#### tag-agentが自動実行

```bash
🏷️ tag-agentのTAGチェーン検証:

## スキャン結果
スキャン対象: .moai/, src/, tests/, docs/
検出されたTAG: 12個

## TAGチェーン整合性検証
✅ @SPEC:EX-HELLO-001 → .moai/specs/SPEC-HELLO-001/spec.md (存在)
✅ @TEST:EX-HELLO-001 → tests/test_hello.py (存在)
✅ @CODE:EX-HELLO-001:API → src/hello/api.py (存在)
✅ @CODE:EX-HELLO-001:MODEL → src/hello/models.py (存在)
✅ @DOC:EX-HELLO-001 → docs/api/hello.md (生成予定)

## Orphan TAG検出
<span class="material-icons">cancel</span> 検出されたOrphan TAG: 0個
✅ すべてのTAGが適切に連結されています

## TAG一貫性検証
✅ すべてのTAGが同じIDを使用 (HELLO-001)
✅ TAG形式が標準に準拠 (@TYPE:EX-DOMAIN-NNN)
```

#### Orphan TAG回復

```
⚙️ Orphan TAG回復処理:

検出された問題:
<span class="material-icons">cancel</span> @CODE:EX-USER-005 (参照先SPECが存在しません)

自動回復アクション:
1. 関連SPEC検索: .moai/specs/ で USER-005 を検索
2. 類似SPEC分析: USER-002, USER-003 と比較
3. 推奨アクション: @CODE:EX-USER-002 にTAG修正

実行しますか？ [Y/n]
```

### フェーズ2: Living Document生成

#### doc-syncerが自動実行

```markdown
# 生成されるLiving Document

# `@DOC:EX-HELLO-001: Hello World API

## 概要

このドキュメントはHello World APIの完全な仕様と実装詳細を説明します。

## 要件

### 機能要件

- システムはHTTP GET /helloエンドポイントを提供すべきである
- WHEN クエリパラメータnameが提供されたら、"Hello, {name}!"を返すべきである
- WHEN nameがない場合、"Hello, World!"を返すべきである

### 非機能要件

- nameは最大50文字に制限すべきである
- 無効な文字が含まれる場合、400エラーを返すべきである
- レスポンスタイムは100ms以内であるべきである

## 実装

### APIエンドポイント

| メソッド | パス | 説明 |
|---------|------|------|
| GET | /hello | 挨拶メッセージを返す |

### リクエストパラメータ

| パラメータ | タイプ | 必須 | 説明 | 制約 |
|-----------|------|------|------|------|
| name | query | いいえ | 挨拶する名前 | 1-50文字、有効な文字のみ |

### レスポンス

#### 成功 (200)

```json
{
  "message": "Hello, 田中!"
}
```

#### エラー (400)

```json
{
  "detail": "Name too long (max 50 chars)"
}
```

## 追跡性

- **要件**: @SPEC:EX-HELLO-001
- **テスト**: @TEST:EX-HELLO-001
- **実装**:
  - @CODE:EX-HELLO-001:API (APIエンドポイント)
  - @CODE:EX-HELLO-001:MODEL (データモデル)
  - @CODE:EX-HELLO-001:SERVICE (ビジネスロジック)
- **ドキュメント**: @DOC:EX-HELLO-001

## テスト

### テストカバレッジ

- 全体カバレッジ: 95%
- APIエンドポイント: 100%
- バリデーション: 100%
- エラーハンドリング: 100%

### テストケース

1. **名前付き挨拶**: nameパラメータでパーソナライズされた挨拶
2. **デフォルト挨拶**: nameなしでデフォルト挨拶
3. **長すぎる名前**: 50文字超えで400エラー
4. **無効な文字**: スクリプト文字で400エラー

## 品質

### TRUST 5原則準拠

✅ **Test First**: 95%カバレッジ達成
✅ **Readable**: クリアなコード構造
✅ **Unified**: 一貫したアーキテクチャ
✅ **Secured**: 入力検証とXSS防止
✅ **Trackable**: 完全なTAG連鎖

## デプロイメント

### 環境変数

```bash
# 開発環境
HELLO_API_DEBUG=true
HELLO_API_LOG_LEVEL=debug

# 本番環境
HELLO_API_DEBUG=false
HELLO_API_LOG_LEVEL=info
```

### ヘルスチェック

```bash
GET /health
{
  "status": "healthy",
  "version": "1.0.0",
  "timestamp": "2025-01-06T10:00:00Z"
}
```
```

### フェーズ3: プロジェクトドキュメント更新

#### README.md自動更新

```markdown
# 更新されるREADME.mdセクション

## 機能

### ✅ Hello World API (HELLO-001)
- パーソナライズされた挨拶メッセージ提供
- クエリパラメータnameをサポート
- 入力検証とエラーハンドリング

**APIエンドポイント**: `GET /hello`
**ドキュメント**: [APIドキュメント](docs/api/hello.md)
**テストカバレッジ**: 95%

### 🏗️ User Authentication (AUTH-001) - 進行中
- JWTベース認証システム
- メール/パスワード認証
- トークンリフレッシュ機能

**ステータス**: 開発中
**進行率**: 75%

## クイックスタート

### Hello API使用例

```bash
# 基本挨拶
curl "http://localhost:8000/hello"
# → {"message": "Hello, World!"}

# パーソナライズされた挨拶
curl "http://localhost:8000/hello?name=田中"
# → {"message": "Hello, 田中!"}
```
```

#### CHANGELOG.md自動更新

```markdown
# 更新されるCHANGELOG.md

# Changelog

## [0.1.0] - 2025-01-06

### Added

#### Hello World API (@SPEC:EX-HELLO-001)
- GET /helloエンドポイント実装
- クエリパラメータnameサポート
- 入力検証（長さ、無効な文字）
- エラーハンドリングと適切なHTTPステータスコード
- 自動APIドキュメンテーション生成

### Implementation Details

- **SPEC**: `.moai/specs/SPEC-HELLO-001/spec.md`
- **Tests**: `tests/test_hello.py` (95% coverage)
- **Code**: `src/hello/` with API, models, and services layers
- **Documentation**: `docs/api/hello.md` with complete API reference

### Quality Metrics

- Test Coverage: 95%
- Code Quality: A+ (TRUST 5 principles compliant)
- Performance: < 10ms response time
- Security: Input validation and XSS prevention

### Contributors

- @user - Implementation and testing

---

## [0.0.1] - 2025-01-05

### Added
- Initial project setup
- MoAI-ADK integration
- Basic configuration
```

### フェーズ4: Git状態管理

#### git-managerが自動実行

```bash
🚀 git-managerの状態管理:

## 変更検出
検出された変更:
  Modified: src/hello/api.py
  Modified: src/hello/models.py
  New: docs/api/hello.md
  Modified: README.md
  Modified: CHANGELOG.md

## コミット推奨
📄 推奨コミットメッセージ:
✅ docs(HELLO-001): sync documentation and update project files

変更内容:
- Living Document生成 (docs/api/hello.md)
- README.md機能セクション更新
- CHANGELOG.mdにv0.1.0リリースノート追加
- TAGチェーン検証完了
```

## 高度な機能

### 自動モード

```bash
/alfred:3-sync auto
```

**自動モード機能**:
- 変更検出時に自動同期実行
- バックグラウンドで定期的実行
- PR作成時に自動実行
- コミット前に自動検証

### 選択的同期

```bash
# 特定SPECのみ同期
/alfred:3-sync HELLO-001

# ドキュメントのみ同期
/alfred:3-sync --docs-only

# TAG検証のみ実行
/alfred:3-sync --tags-only

# レポートのみ生成
/alfred:3-sync --report-only
```

### カスタムテンプレート

```yaml
# .moai/templates/sync-custom.yml
custom_documentation:
  enabled: true
  template: "custom-api-doc.md"
  output: "docs/custom/{SPEC_ID}.md"

report_format:
  format: "json"
  include_metrics: true
  include_recommendations: true
```

## 同期レポート

### レポート生成

```bash
/alfred:3-sync --report
```

### レポート内容

```
📊 同期レポート (2025-01-06 10:00:00)

## 実行概要
- 実行時間: 45秒
- 処理ファイル: 12個
- 生成ドキュメント: 3個
- 更新ファイル: 5個

## 品質メトリクス
- 全体カバレッジ: 95% ↗️ (+2%)
- TRUST準拠率: 100%
- TAG整合性: 100%
- ドキュメント最新性: 100%

## 検出された問題
<span class="material-icons">cancel</span> 警告: src/unused.py (未使用ファイル)
<span class="material-icons">cancel</span> 警告: README.md (APIドキュメントリンク切れ)

## 推奨アクション
1. 未使用ファイルを削除または移動
2. README.mdのリンクを更新
3. 次のリリース準備完了

## 次のステップ
✅ プルリクエスト作成準備完了
✅ コミット推奨
✅ デプロイ可能な状態
```

### レポート出力形式

```bash
# JSON形式
/alfred:3-sync --report --format=json

# HTML形式
/alfred:3-sync --report --format=html

# Markdown形式（デフォルト）
/alfred:3-sync --report --format=markdown
```

## 状態遷移管理

### SPEC状態更新

```bash
# 状態確認
grep "status:" .moai/specs/SPEC-HELLO-001/spec.md
# 出力: status: in_progress

# 同期後状態更新
/alfred:3-sync

# 状態確認
grep "status:" .moai/specs/SPEC-HELLO-001/spec.md
# 出力: status: completed
```

### ワークフロー統合

```mermaid
%%{init: {'theme':'neutral'}}%%
graph TD
    Plan[/alfred:1-plan<br/>status: planning] --> Run[/alfred:2-run<br/>status: in_progress]
    Run --> Sync[/alfred:3-sync<br/>status: testing]
    Sync --> Completed[status: completed]
    Completed --> Plan

    Sync -.-> PR[Pull Request<br/>Ready for Review]
    PR -.-> Merge[Merge<br/>status: stable]
```

## ドキュメントタイプ

### 1. APIドキュメント

自動生成されるAPIリファレンス：

```markdown
# APIドキュメント構造

## エンドポイント一覧
- パス、メソッド、説明
- パラメータ詳細
- レスポンス形式
- エラーコード

## スキーマ定義
- リクエスト/レスポンスモデル
- データ型と制約
- バリデーションルール

## 使用例
- cURLコマンド例
- レスポンス例
- エラーケース例
```

### 2. アーキテクチャドキュメント

```markdown
# アーキテクチャ概要

## システム構成
- コンポーネント図
- データフロー
- 技術スタック

## 設計決定
- 選択した技術と理由
- トレードオフ分析
- 将来の拡張性計画
```

### 3. デプロイドキュメント

```markdown
# デプロイガイド

## 環境要件
- 依存関係リスト
- システム要件
- 設定パラメータ

## デプロイ手順
- ステップバイステップガイド
- 環境変数設定
- ヘルスチェック
```

## 品質保証

### ドキュメント品質検証

```bash
📚 doc-syncerの品質検証:

## ドキュメント完全性
✅ すべての@SPECに対応する@DOCが存在
✅ すべての@CODEにドキュメント化済み
✅ すべての@TESTに使用例が記載

## ドキュメント一貫性
✅ 用語と命名規則の一貫性
✅ フォーマットと構造の統一性
✅ バージョン情報の正確性

## ドキュメントアクセシビリティ
✅ 見出しと目次の完成度
✅ コード例の実行可能性
✅ 外部リンクの有効性
```

### リンク検証

```bash
# 内部リンク検証
/alfred:3-sync --validate-links

# 外部リンク検証
/alfred:3-sync --validate-external-links

# 画像リソース検証
/alfred:3-sync --validate-images
```

## トラブルシューティング

### よくある問題

**ドキュメントが生成されない**:
```bash
# TAGチェーン確認
rg '@(SPEC|TEST|CODE):' -n

# 同期強制実行
/alfred:3-sync --force

# 詳細ログ確認
/alfred:3-sync --verbose
```

**Orphan TAGが多い**:
```bash
# Orphan TAG分析
/alfred:3-sync --analyze-orphans

# 自動修復提案
/alfred:3-sync --suggest-fixes
```

**ドキュメントフォーマットが壊れる**:
```bash
# テンプレート再生成
/alfred:3-sync --regenerate-templates

# Markdown検証
/alfred:3-sync --validate-markdown
```

## ベストプラクティス

### 1. 定期的な同期

```bash
# 開発サイクル終了後必ず実行
/alfred:2-run SPEC-001
/alfred:3-sync  # 忘れずに実行

# コミット前の最終確認
git add .
/alfred:3-sync --report-only  # レポート確認
git commit -m "message"
```

### 2. ドキュメント品質維持

- **コード変更と同時ドキュメント更新**
- **定期的なリンク検証**
- **用語集の維持**
- **バージョン管理の徹底**

### 3. チーム協業

```bash
# ドキュメントレビュープロセス
1. /alfred:3-sync 実行
2. 生成されたドキュメントをチームレビュー
3. フィードバック反映
4. 再同期実行
5. レビュー完了
```

## 統合と連携

### CI/CDパイプライン統合

```yaml
# .github/workflows/sync.yml
name: Documentation Sync
on:
  push:
    branches: [main, develop]
  pull_request:
    branches: [main]

jobs:
  sync:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - name: Setup Python
        uses: actions/setup-python@v4
        with:
          python-version: '3.13'
      - name: Install MoAI-ADK
        run: pip install moai-adk
      - name: Run sync
        run: /alfred:3-sync --report --format=json
      - name: Validate documentation
        run: |
          # ドキュメント品質検証
          # リンク検証
          # カバレッジ確認
```

### デプロイメント連携

```bash
# 本番環境展開前
/alfred:3-sync --production-ready

# デプロイ後
/alfred:3-sync --update-deployment-status
```

---

**📚 次のステップ**:
- [プロジェクトガイド](../project/index.md)でプロジェクト管理
- [デプロイガイド](../project/deploy.md)で本番環境展開
- [品質ガイド](../project/config.md)で品質管理